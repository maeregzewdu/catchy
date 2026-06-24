package mail

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/google/uuid"

	"github.com/maeregzewdu/catchy/internal/model"
	"github.com/maeregzewdu/catchy/internal/store"
)

// SyncAccount connects to the account's IMAP server and incrementally fetches
// new messages from each configured folder. Per-folder failures are logged and
// skipped; only connection/login errors are fatal.
func SyncAccount(ctx context.Context, a *model.Account, s store.Store, dataDir string, folders []string) error {
	c, err := dialIMAP(a)
	if err != nil {
		return fmt.Errorf("IMAP connect: %w", err)
	}
	defer c.Close()

	if err := c.Login(a.Username, a.Password).Wait(); err != nil {
		return fmt.Errorf("LOGIN: %w", err)
	}
	defer c.Logout().Wait() //nolint:errcheck

	for _, folder := range folders {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := syncFolder(ctx, c, a, s, dataDir, folder); err != nil {
			slog.Warn("sync folder", "account", a.Email, "folder", folder, "err", err)
		}
	}
	return nil
}

func syncFolder(ctx context.Context, c *imapclient.Client, a *model.Account, s store.Store, dataDir, folder string) error {
	state, err := s.GetSyncState(a.ID, folder)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return err
	}

	var localUIDNext uint32 = 1
	var localUIDValidity uint32
	if state != nil {
		localUIDNext = state.UIDNext
		localUIDValidity = state.UIDValidity
	}

	mboxData, err := c.Select(folder, nil).Wait()
	if err != nil {
		return fmt.Errorf("SELECT %s: %w", folder, err)
	}

	serverUIDValidity := mboxData.UIDValidity

	// UIDVALIDITY change means all cached UIDs are invalid — start fresh.
	if localUIDValidity != 0 && serverUIDValidity != localUIDValidity {
		slog.Info("UIDVALIDITY changed, re-syncing from scratch",
			"account", a.Email, "folder", folder)
		localUIDNext = 1
	}

	serverUIDNext := uint32(mboxData.UIDNext)
	if serverUIDNext <= localUIDNext {
		now := time.Now().UTC()
		return s.UpsertSyncState(&model.SyncState{
			AccountID:   a.ID,
			Folder:      folder,
			UIDNext:     localUIDNext,
			UIDValidity: serverUIDValidity,
			LastSync:    &now,
		})
	}

	// Fetch UIDs [localUIDNext, *] with full body (PEEK to avoid marking Seen).
	var uidSet imap.UIDSet
	uidSet.AddRange(imap.UID(localUIDNext), 0) // stop=0 means *

	sectionSpec := &imap.FetchItemBodySection{Peek: true}
	messages, err := c.Fetch(uidSet, &imap.FetchOptions{
		UID:          true,
		Flags:        true,
		InternalDate: true,
		BodySection:  []*imap.FetchItemBodySection{sectionSpec},
	}).Collect()
	if err != nil {
		return fmt.Errorf("UID FETCH: %w", err)
	}

	var maxUID uint32 = localUIDNext - 1

	for _, msgBuf := range messages {
		if ctx.Err() != nil {
			break
		}

		uid := uint32(msgBuf.UID)
		if uid > maxUID {
			maxUID = uid
		}

		var raw []byte
		if len(msgBuf.BodySection) > 0 {
			raw = msgBuf.BodySection[0].Bytes
		}
		if len(raw) == 0 {
			continue
		}

		msgID := uuid.New().String()
		msg, atts, err := parseMIME(raw, dataDir, msgID)
		if err != nil {
			slog.Warn("parse MIME", "uid", uid, "err", err)
			// parseMIME returns a partial msg even on error; continue storing it.
		}

		accountID := a.ID
		msg.AccountID = &accountID
		msg.Source = sourceForFolder(folder)
		msg.Folder = folder
		msg.IsRead = hasFlag(msgBuf.Flags, imap.FlagSeen)
		msg.IsStarred = hasFlag(msgBuf.Flags, imap.FlagFlagged)
		t := msgBuf.InternalDate
		msg.ReceivedAt = &t

		if err := s.CreateMessage(msg, atts); err != nil {
			slog.Warn("store message", "uid", uid, "err", err)
		}
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	now := time.Now().UTC()
	return s.UpsertSyncState(&model.SyncState{
		AccountID:   a.ID,
		Folder:      folder,
		UIDNext:     maxUID + 1,
		UIDValidity: serverUIDValidity,
		LastSync:    &now,
	})
}

// dialIMAP opens a connection appropriate for the account's IMAP port:
// 993 → implicit TLS, 143 → STARTTLS, anything else → plain.
func dialIMAP(a *model.Account) (*imapclient.Client, error) {
	addr := net.JoinHostPort(a.IMAPHost, strconv.Itoa(a.IMAPPort))
	opts := &imapclient.Options{
		TLSConfig: &tls.Config{ServerName: a.IMAPHost},
	}
	switch a.IMAPPort {
	case 993:
		return imapclient.DialTLS(addr, opts)
	case 143:
		return imapclient.DialStartTLS(addr, opts)
	default:
		return imapclient.DialInsecure(addr, nil)
	}
}

func sourceForFolder(folder string) string {
	if strings.Contains(strings.ToLower(folder), "sent") {
		return "sent"
	}
	return "imap"
}

func hasFlag(flags []imap.Flag, target imap.Flag) bool {
	for _, f := range flags {
		if f == target {
			return true
		}
	}
	return false
}
