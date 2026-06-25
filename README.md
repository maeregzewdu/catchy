# Catchy

A local email dev tool for catching outgoing mail from your applications.

**Trap inbox** — spin up a zero-config SMTP server on `localhost:1025`. Every message your app sends lands in a web UI instantly. No real emails sent, no configuration on the mail server side.

**Real accounts** — connect your staging Gmail, company mail server, or any IMAP/SMTP account to read and inspect actual production-ish mail.

---

## Quickstart

### Prerequisites

- Go 1.22+

### Build and run

```bash
git clone https://github.com/maeregzewdu/catchy
cd catchy
go build -o catchy ./cmd/catchy
./catchy serve
```

Open `http://localhost:8080` in your browser.

The SMTP trap listens on `localhost:1025` by default. No auth required.

### Seed demo data

```bash
./catchy seed
```

Adds 15 realistic trap messages and two demo IMAP accounts with ~20 messages each. Use `--reset` to clear trap messages before seeding:

```bash
./catchy seed --reset
```

---

## Configuration

Catchy looks for a config file at `~/.catchy/config.toml`. It is created automatically on first run with sensible defaults.

```toml
[server]
host = "127.0.0.1"
port = 8080

[trap]
host = "127.0.0.1"
port = 1025

[data]
dir = "~/.catchy/data"

[sync]
poll_interval_seconds = 60
default_folders = ["INBOX", "Sent"]
```

---

## Connecting your app to the trap

Point your app's SMTP settings at `localhost:1025`. No password needed.

### Laravel

```env
MAIL_MAILER=smtp
MAIL_HOST=127.0.0.1
MAIL_PORT=1025
MAIL_USERNAME=null
MAIL_PASSWORD=null
MAIL_ENCRYPTION=null
```

### Node.js (nodemailer)

```js
const transporter = nodemailer.createTransport({
  host: '127.0.0.1',
  port: 1025,
  secure: false,
  ignoreTLS: true,
});
```

### swaks (CLI testing)

```bash
swaks --to test@example.com --from sender@dev.local \
  --server 127.0.0.1:1025 \
  --body "Hello from the trap"
```

---

## Adding a real mail account

Go to **Settings** in the web UI and fill in the IMAP/SMTP credentials. Gmail users should create an [App Password](https://support.google.com/accounts/answer/185833) and use port 587 (SMTP) / 993 (IMAP).

Catchy syncs the folders listed in `sync.default_folders` every `poll_interval_seconds` seconds, and you can trigger a manual sync from the folder view.

---

## API

The JSON API lives at `/api/v1`:

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/health` | Health check |
| `GET` | `/api/v1/trap/messages` | List trap messages |
| `GET` | `/api/v1/trap/messages/:id` | Get trap message + attachments |
| `PATCH` | `/api/v1/trap/messages/:id` | Star / mark read |
| `DELETE` | `/api/v1/trap/messages/:id` | Delete one trap message |
| `DELETE` | `/api/v1/trap/messages` | Clear all trap messages |
| `GET` | `/api/v1/search?q=` | Full-text search (FTS5) |
| `GET` | `/api/v1/accounts` | List accounts |
| `POST` | `/api/v1/accounts` | Create account |
| `POST` | `/api/v1/accounts/:id/verify` | Test SMTP + IMAP credentials |
| `POST` | `/api/v1/accounts/:id/sync` | Trigger IMAP sync |
| `GET` | `/api/v1/accounts/:id/messages?folder=` | List messages |
| `PATCH` | `/api/v1/accounts/:id/messages/:msgId` | Star / mark read |
| `DELETE` | `/api/v1/accounts/:id/messages/:msgId` | Delete message |

---

## Single binary

All HTML templates are embedded in the binary via `//go:embed`. The built binary is self-contained:

```bash
go build -ldflags="-s -w" -o catchy ./cmd/catchy
```
