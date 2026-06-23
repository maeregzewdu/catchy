package store

import (
	"errors"

	"github.com/maeregzewdu/catchy/internal/model"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrNotImplemented = errors.New("not implemented")
)

// TrapFilter constrains ListTrapMessages results. Empty fields are ignored.
type TrapFilter struct {
	To      string
	From    string
	Subject string
}

// MessageFilter constrains ListMessages results. Nil pointer fields are ignored.
type MessageFilter struct {
	Read    *bool
	Starred *bool
}

// Store is the single persistence interface for all catchy data.
// Methods used in later phases are stubbed and return ErrNotImplemented until implemented.
type Store interface {
	// Health
	Ping() error
	Close() error

	// Trap inbox — Phase 2
	StoreTrapMessage(msg *model.Message, attachments []model.Attachment) error
	ListTrapMessages(filters TrapFilter) ([]*model.Message, error)
	GetTrapMessage(id string) (*model.Message, error)
	DeleteTrapMessage(id string) error
	ClearTrapMessages() error

	// Accounts — Phase 3
	CreateAccount(a *model.Account) error
	GetAccount(id string) (*model.Account, error)
	ListAccounts() ([]*model.Account, error)
	UpdateAccount(a *model.Account) error
	DeleteAccount(id string) error

	// Messages (real accounts) — Phase 4
	CreateMessage(msg *model.Message) error
	GetMessage(id string) (*model.Message, error)
	ListMessages(accountID, folder string, filters MessageFilter) ([]*model.Message, error)
	UpdateMessage(msg *model.Message) error
	PatchMessage(id string, read *bool, starred *bool) error
	DeleteMessage(id string) error
	GetAttachment(id string) (*model.Attachment, error)
	ListAttachments(messageID string) ([]model.Attachment, error)

	// IMAP sync state — Phase 4
	GetSyncState(accountID, folder string) (*model.SyncState, error)
	UpsertSyncState(state *model.SyncState) error

	// Full-text search — Phase 6
	SearchMessages(q, accountID, source string) ([]*model.Message, error)
}
