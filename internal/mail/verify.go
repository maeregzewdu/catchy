package mail

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/maeregzewdu/catchy/internal/model"
)

const verifyTimeout = 10 * time.Second

// VerifySMTP tests SMTP connectivity and AUTH for the account.
// It selects implicit TLS (port 465) or STARTTLS (all other ports) automatically.
func VerifySMTP(a *model.Account) error {
	addr := net.JoinHostPort(a.SMTPHost, strconv.Itoa(a.SMTPPort))
	dialer := &net.Dialer{Timeout: verifyTimeout}

	var c *smtp.Client

	if a.SMTPPort == 465 {
		conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: a.SMTPHost}) //nolint:govet
		if err != nil {
			return fmt.Errorf("TLS connect: %w", err)
		}
		c, err = smtp.NewClient(conn, a.SMTPHost)
		if err != nil {
			conn.Close()
			return fmt.Errorf("SMTP handshake: %w", err)
		}
	} else {
		conn, err := dialer.Dial("tcp", addr)
		if err != nil {
			return fmt.Errorf("connect: %w", err)
		}
		c, err = smtp.NewClient(conn, a.SMTPHost)
		if err != nil {
			conn.Close()
			return fmt.Errorf("SMTP handshake: %w", err)
		}
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(&tls.Config{ServerName: a.SMTPHost}); err != nil {
				c.Close()
				return fmt.Errorf("STARTTLS: %w", err)
			}
		}
	}
	defer c.Quit() //nolint:errcheck

	if err := c.Auth(smtp.PlainAuth("", a.Username, a.Password, a.SMTPHost)); err != nil {
		return fmt.Errorf("AUTH: %w", err)
	}
	return nil
}

// VerifyIMAP tests IMAP connectivity and LOGIN for the account using a raw
// TCP session. This avoids pulling in the full go-imap library for a simple
// credential check; Phase 4 uses go-imap for ongoing sync.
func VerifyIMAP(a *model.Account) error {
	addr := net.JoinHostPort(a.IMAPHost, strconv.Itoa(a.IMAPPort))
	dialer := &net.Dialer{Timeout: verifyTimeout}

	var conn net.Conn
	var err error

	if a.IMAPPort == 993 {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: a.IMAPHost})
	} else {
		conn, err = dialer.Dial("tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(verifyTimeout)) //nolint:errcheck

	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	// Read server greeting.
	greeting, err := r.ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading greeting: %w", err)
	}
	if !strings.HasPrefix(greeting, "* OK") && !strings.HasPrefix(greeting, "* PREAUTH") {
		return fmt.Errorf("unexpected greeting: %s", strings.TrimSpace(greeting))
	}

	// Send LOGIN command.
	fmt.Fprintf(w, "A1 LOGIN %s %s\r\n", imapQuote(a.Username), imapQuote(a.Password))
	if err := w.Flush(); err != nil {
		return fmt.Errorf("sending LOGIN: %w", err)
	}

	// Read until we get the tagged response (A1 OK / A1 NO / A1 BAD).
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading LOGIN response: %w", err)
		}
		if strings.HasPrefix(line, "A1 OK") {
			break
		}
		if strings.HasPrefix(line, "A1 NO") || strings.HasPrefix(line, "A1 BAD") {
			return fmt.Errorf("auth failed: %s", strings.TrimSpace(line))
		}
		// Untagged (* ...) lines — keep reading.
	}

	// Logout cleanly; ignore errors — the auth check already passed.
	fmt.Fprint(w, "A2 LOGOUT\r\n")
	w.Flush() //nolint:errcheck
	return nil
}

// imapQuote wraps s in double quotes if it contains characters that require
// quoting in an IMAP atom.
func imapQuote(s string) string {
	if !strings.ContainsAny(s, " \"\\{}\r\n") {
		return s
	}
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}
