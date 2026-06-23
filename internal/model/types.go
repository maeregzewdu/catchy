package model

import "time"

type Account struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	SMTPHost  string    `json:"smtp_host"`
	SMTPPort  int       `json:"smtp_port"`
	IMAPHost  string    `json:"imap_host"`
	IMAPPort  int       `json:"imap_port"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // never marshalled to JSON
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	ID         string     `json:"id"`
	AccountID  *string    `json:"account_id,omitempty"` // nil for trap messages
	Source     string     `json:"source"`               // "trap" | "imap" | "sent"
	MessageID  string     `json:"message_id,omitempty"`
	Folder     string     `json:"folder"`
	Subject    string     `json:"subject,omitempty"`
	FromAddr   string     `json:"from_addr"`
	ToAddrs    []string   `json:"to_addrs"`
	CCAddrs    []string   `json:"cc_addrs,omitempty"`
	BCCAddrs   []string   `json:"bcc_addrs,omitempty"`
	BodyText   string     `json:"body_text,omitempty"`
	BodyHTML   string     `json:"body_html,omitempty"`
	RawMIME    []byte     `json:"-"`
	IsRead     bool       `json:"is_read"`
	IsStarred  bool       `json:"is_starred"`
	SentAt     *time.Time `json:"sent_at,omitempty"`
	ReceivedAt *time.Time `json:"received_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type Attachment struct {
	ID        string `json:"id"`
	MessageID string `json:"message_id"`
	Filename  string `json:"filename"`
	MIMEType  string `json:"mime_type"`
	Size      int64  `json:"size"`
	Path      string `json:"path"`
}

type SyncState struct {
	AccountID   string     `json:"account_id"`
	Folder      string     `json:"folder"`
	UIDNext     uint32     `json:"uid_next"`
	UIDValidity uint32     `json:"uid_validity"`
	LastSync    *time.Time `json:"last_sync,omitempty"`
}
