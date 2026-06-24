package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/maeregzewdu/catchy/internal/config"
	"github.com/maeregzewdu/catchy/internal/model"
	"github.com/maeregzewdu/catchy/internal/store"
)

func runSeed(args []string) error {
	fs := flag.NewFlagSet("seed", flag.ExitOnError)
	reset := fs.Bool("reset", false, "clear existing trap messages and demo accounts before seeding")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	if err := os.MkdirAll(cfg.Data.Dir, 0755); err != nil {
		return fmt.Errorf("creating data dir: %w", err)
	}

	catchyDir := config.DefaultDir()
	key, err := config.LoadOrCreateSecretKey(catchyDir)
	if err != nil {
		return fmt.Errorf("secret key: %w", err)
	}
	db, err := store.New(cfg.Data.Dir, key)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	if *reset {
		slog.Info("clearing existing trap messages")
		if err := db.ClearTrapMessages(); err != nil {
			slog.Warn("clear trap", "err", err)
		}
	}

	slog.Info("seeding trap messages")
	if err := seedTrap(db); err != nil {
		return fmt.Errorf("seed trap: %w", err)
	}

	slog.Info("seeding demo accounts and IMAP messages")
	if err := seedAccounts(db); err != nil {
		return fmt.Errorf("seed accounts: %w", err)
	}

	slog.Info("seed complete — run 'catchy serve' and open http://localhost:8080")
	return nil
}

// ── Trap messages ─────────────────────────────────────────────────────────────

type seedMsg struct {
	from    string
	to      string
	subject string
	text    string
	html    string
	agoMin  int // received N minutes ago
	read    bool
	starred bool
}

func seedTrap(db store.Store) error {
	now := time.Now()
	msgs := []seedMsg{
		{
			from: "no-reply@myapp.dev", to: "test@codingbz.com",
			subject: "Welcome to MyApp — please verify your email",
			agoMin:  2, starred: true,
			html: `<div style="font-family:sans-serif;max-width:600px;margin:0 auto;padding:24px">
<h2 style="color:#1e40af">Welcome to MyApp!</h2>
<p>Hi there, thanks for signing up. Please verify your email address to get started.</p>
<a href="#" style="display:inline-block;background:#2563eb;color:#fff;padding:12px 24px;border-radius:6px;text-decoration:none;font-weight:600;margin:16px 0">Verify Email Address</a>
<p style="color:#6b7280;font-size:14px">If you didn't create this account, you can safely ignore this email.</p>
</div>`,
		},
		{
			from: "no-reply@myapp.dev", to: "test@codingbz.com",
			subject: "Your password reset link",
			agoMin:  15,
			html: `<div style="font-family:sans-serif;max-width:600px;margin:0 auto;padding:24px">
<h2 style="color:#dc2626">Password Reset Request</h2>
<p>We received a request to reset your password. Click the button below to choose a new password.</p>
<a href="#" style="display:inline-block;background:#dc2626;color:#fff;padding:12px 24px;border-radius:6px;text-decoration:none;font-weight:600;margin:16px 0">Reset Password</a>
<p style="color:#6b7280;font-size:14px">This link expires in 1 hour. If you didn't request a password reset, ignore this email.</p>
</div>`,
		},
		{
			from: "orders@myapp.dev", to: "test@codingbz.com",
			subject: "Order #1042 confirmed — ships in 2–3 business days",
			agoMin:  45, read: true,
			html: `<div style="font-family:sans-serif;max-width:600px;margin:0 auto;padding:24px">
<h2 style="color:#059669">Order Confirmed ✓</h2>
<p>Thanks for your order! Here's a summary:</p>
<table style="width:100%;border-collapse:collapse;margin:16px 0">
<tr style="border-bottom:1px solid #e5e7eb"><td style="padding:8px;font-weight:600">Pro Plan (monthly)</td><td style="padding:8px;text-align:right">$29.00</td></tr>
<tr><td style="padding:8px;font-weight:600;color:#374151">Total</td><td style="padding:8px;text-align:right;font-weight:600">$29.00</td></tr>
</table>
<p style="color:#6b7280;font-size:14px">Order #1042 · Payment via Stripe</p>
</div>`,
		},
		{
			from: "notifications@myapp.dev", to: "test@codingbz.com",
			subject: "[Alert] High CPU usage on prod-server-1 (94%)",
			agoMin:  90, starred: true,
			text:    "ALERT: prod-server-1 CPU usage at 94% for the last 10 minutes.\n\nHost: prod-server-1.us-east-1\nMetric: cpu_usage\nThreshold: 90%\nCurrent: 94%\nStarted: 2024-01-15 14:22:01 UTC\n\nDashboard: https://metrics.example.com/servers/prod-1\n\nThis is an automated alert from MyApp monitoring.",
		},
		{
			from: "admin@myapp.dev", to: "test@codingbz.com",
			subject: "Weekly digest: 1,247 new signups this week",
			agoMin:  180, read: true,
			html: `<div style="font-family:sans-serif;max-width:600px;margin:0 auto;padding:24px">
<h2>Weekly Digest — Jan 8–14, 2024</h2>
<div style="display:grid;grid-template-columns:1fr 1fr 1fr;gap:16px;margin:16px 0">
<div style="background:#f0f9ff;padding:16px;border-radius:8px;text-align:center"><div style="font-size:24px;font-weight:700;color:#0369a1">1,247</div><div style="color:#6b7280;font-size:13px">New Signups</div></div>
<div style="background:#f0fdf4;padding:16px;border-radius:8px;text-align:center"><div style="font-size:24px;font-weight:700;color:#15803d">$8,940</div><div style="color:#6b7280;font-size:13px">Revenue</div></div>
<div style="background:#fef9c3;padding:16px;border-radius:8px;text-align:center"><div style="font-size:24px;font-weight:700;color:#a16207">3.2%</div><div style="color:#6b7280;font-size:13px">Churn Rate</div></div>
</div>
</div>`,
		},
		{
			from: "noreply@github.com", to: "test@codingbz.com",
			subject: "[myapp/backend] PR #47: Add rate limiting middleware",
			agoMin:  240, read: true,
			text:    "maeregzewdu opened a pull request.\n\nTitle: Add rate limiting middleware\nBranch: feat/rate-limit → main\n\nDescribes adding per-IP rate limiting using a sliding window algorithm.\n\nView PR: https://github.com/myapp/backend/pull/47",
		},
		{
			from: "no-reply@myapp.dev", to: "qa@codingbz.com",
			subject: "Test email — HTML + plain text multipart",
			agoMin:  300, read: true,
			text: "This is the plain text version of the email.",
			html: `<div style="font-family:monospace;padding:16px;background:#1e1e1e;color:#d4d4d4;border-radius:8px">
<p style="color:#4ec9b0">// Test email for UI development</p>
<p style="color:#9cdcfe">const</p> <span style="color:#4fc1ff">subject</span> = <span style="color:#ce9178">"HTML + plain text multipart"</span>;<br>
<p style="color:#9cdcfe">const</p> <span style="color:#4fc1ff">hasHTML</span> = <span style="color:#569cd6">true</span>;
</div>`,
		},
		{
			from: "billing@stripe.com", to: "test@codingbz.com",
			subject: "Your Stripe invoice is ready — $149.00",
			agoMin:  400,
			html: `<div style="font-family:sans-serif;max-width:600px;margin:0 auto;padding:24px">
<div style="display:flex;align-items:center;gap:12px;margin-bottom:24px">
<div style="width:32px;height:32px;background:#635bff;border-radius:6px"></div>
<strong style="font-size:18px">Stripe</strong>
</div>
<h2>Invoice from Acme Corp</h2>
<p>Amount due: <strong>$149.00</strong></p>
<p>Due date: February 1, 2024</p>
<a href="#" style="display:inline-block;background:#635bff;color:#fff;padding:10px 20px;border-radius:6px;text-decoration:none;font-weight:500;margin:12px 0">Pay Invoice</a>
</div>`,
		},
		{
			from: "deploy@vercel.com", to: "test@codingbz.com",
			subject: "✅ Deployment successful: myapp.vercel.app",
			agoMin:  600, read: true,
			text: "Your deployment is live!\n\nProject: myapp\nBranch: main\nCommit: a3f8c12 (Add dark mode toggle)\nURL: https://myapp.vercel.app\nBuild time: 42s\n\nView deployment: https://vercel.com/dashboard",
		},
		{
			from: "no-reply@linear.app", to: "test@codingbz.com",
			subject: "ENG-234 assigned to you: Fix memory leak in sync loop",
			agoMin:  720,
			html: `<div style="font-family:sans-serif;max-width:600px;margin:0 auto;padding:24px">
<p style="color:#6b7280;font-size:13px">Linear · ENG-234</p>
<h2 style="margin-top:4px">Fix memory leak in sync loop</h2>
<p>This issue has been assigned to you by <strong>alex</strong>.</p>
<div style="background:#f8fafc;border:1px solid #e2e8f0;border-radius:8px;padding:12px;margin:12px 0">
<p style="margin:0;font-size:14px;color:#374151">The IMAP sync loop appears to hold references to closed connections. Heap profiling shows ~200MB growth per hour under load.</p>
</div>
<a href="#" style="color:#6366f1;text-decoration:none;font-weight:500">View in Linear →</a>
</div>`,
		},
		{
			from: "noreply@aws.amazon.com", to: "test@codingbz.com",
			subject: "AWS Free Tier Usage Alert: EC2 approaching limit",
			agoMin:  900, read: true,
			text:    "AWS Free Tier Alert\n\nService: Amazon EC2\nUsage: 715 of 750 free hours used this month.\n\nYou are approaching your monthly Free Tier limit. Additional usage will be charged at standard rates.\n\nRegion: us-east-1\nInstance Type: t2.micro\n\nManage your Free Tier usage in the AWS Console.",
		},
		{
			from: "team@postmark.app", to: "dev@codingbz.com",
			subject: "Transactional email delivery report — January 2024",
			agoMin:  1200,
			html: `<div style="font-family:sans-serif;max-width:600px;margin:0 auto;padding:24px">
<h2>January 2024 Delivery Report</h2>
<p>Here's how your emails performed last month:</p>
<table style="width:100%;border-collapse:collapse">
<tr style="background:#f8fafc"><td style="padding:10px;font-weight:600">Sent</td><td style="padding:10px;text-align:right">12,483</td></tr>
<tr><td style="padding:10px;font-weight:600">Delivered</td><td style="padding:10px;text-align:right">12,391 (99.3%)</td></tr>
<tr style="background:#f8fafc"><td style="padding:10px;font-weight:600">Bounced</td><td style="padding:10px;text-align:right;color:#dc2626">92 (0.7%)</td></tr>
<tr><td style="padding:10px;font-weight:600">Opened</td><td style="padding:10px;text-align:right">4,821 (38.6%)</td></tr>
</table>
</div>`,
		},
		{
			from: "no-reply@myapp.dev", to: "alice@example.com",
			subject: "Magic link — sign in to MyApp",
			agoMin:  1440, read: true,
			html: `<div style="font-family:sans-serif;max-width:560px;margin:0 auto;padding:32px">
<h2>Your sign-in link</h2>
<p>Click the button below to sign in. This link expires in 15 minutes.</p>
<a href="#" style="display:inline-block;background:#0f172a;color:#f8fafc;padding:14px 28px;border-radius:8px;text-decoration:none;font-weight:600;letter-spacing:0.025em;margin:16px 0">Sign in to MyApp</a>
<p style="color:#94a3b8;font-size:13px">Requested from IP 203.0.113.42. If this wasn't you, you can safely ignore this email.</p>
</div>`,
		},
		{
			from: "notifications@myapp.dev", to: "test@codingbz.com",
			subject: "New comment on your post: \"Building a local mail tool\"",
			agoMin:  2000,
			text:    "alex_dev commented on your post:\n\n\"Great writeup! I've been looking for something exactly like this. Does it support OAuth2 for Gmail accounts?\"\n\nView comment: https://myapp.dev/posts/local-mail-tool#comment-892",
		},
		{
			from: "security@myapp.dev", to: "test@codingbz.com",
			subject: "New sign-in from Chrome on Windows",
			agoMin:  2880, read: true, starred: true,
			html: `<div style="font-family:sans-serif;max-width:560px;margin:0 auto;padding:24px">
<div style="background:#fef2f2;border:1px solid #fecaca;border-radius:8px;padding:16px;margin-bottom:16px">
<h3 style="color:#991b1b;margin:0 0 8px">Security Alert</h3>
<p style="margin:0;color:#7f1d1d;font-size:14px">A new device signed into your account.</p>
</div>
<p><strong>Browser:</strong> Chrome 120</p>
<p><strong>OS:</strong> Windows 11</p>
<p><strong>IP:</strong> 203.0.113.42</p>
<p><strong>Location:</strong> Addis Ababa, ET</p>
<p><strong>Time:</strong> Jan 15, 2024 at 9:14 AM</p>
<p>If this was you, no action is needed. If not, <a href="#" style="color:#dc2626">secure your account</a>.</p>
</div>`,
		},
	}

	for _, sm := range msgs {
		receivedAt := now.Add(-time.Duration(sm.agoMin) * time.Minute)
		id := uuid.New().String()
		raw := buildRawMIME(sm.from, sm.to, sm.subject, sm.text, sm.html, receivedAt)
		msg := &model.Message{
			ID:         id,
			Source:     "trap",
			Folder:     "trap",
			MessageID:  "<" + uuid.New().String() + "@seed.catchy>",
			Subject:    sm.subject,
			FromAddr:   sm.from,
			ToAddrs:    []string{sm.to},
			BodyText:   sm.text,
			BodyHTML:   sm.html,
			IsRead:     sm.read,
			IsStarred:  sm.starred,
			RawMIME:    raw,
			ReceivedAt: &receivedAt,
			CreatedAt:  receivedAt,
		}
		if err := db.StoreTrapMessage(msg, nil); err != nil {
			slog.Warn("seed trap message", "subject", sm.subject, "err", err)
		}
	}
	slog.Info("trap messages seeded", "count", len(msgs))
	return nil
}

// ── Demo accounts + IMAP messages ─────────────────────────────────────────────

func seedAccounts(db store.Store) error {
	now := time.Now()

	accounts := []model.Account{
		{
			Name:     "Staging",
			Email:    "staging@myapp.dev",
			SMTPHost: "smtp.gmail.com",
			SMTPPort: 587,
			IMAPHost: "imap.gmail.com",
			IMAPPort: 993,
			Username: "staging@myapp.dev",
			Password: "demo-app-password",
		},
		{
			Name:     "Company",
			Email:    "dev@company.local",
			SMTPHost: "mail.company.local",
			SMTPPort: 465,
			IMAPHost: "mail.company.local",
			IMAPPort: 993,
			Username: "dev",
			Password: "demo-pass",
		},
	}

	for i := range accounts {
		a := &accounts[i]
		existing, _ := db.ListAccounts()
		skip := false
		for _, e := range existing {
			if e.Email == a.Email {
				a.ID = e.ID
				skip = true
				break
			}
		}
		if !skip {
			if err := db.CreateAccount(a); err != nil {
				slog.Warn("seed account", "email", a.Email, "err", err)
				continue
			}
		}
		slog.Info("seeding IMAP messages", "account", a.Email)
		if err := seedIMAPMessages(db, a, now); err != nil {
			slog.Warn("seed imap", "account", a.Email, "err", err)
		}
	}
	return nil
}

type imapSeedMsg struct {
	from    string
	subject string
	text    string
	html    string
	folder  string
	agoMin  int
	read    bool
	starred bool
}

func seedIMAPMessages(db store.Store, a *model.Account, now time.Time) error {
	messages := []imapSeedMsg{
		// INBOX
		{
			folder: "INBOX", from: "noreply@github.com",
			subject: "[myapp/backend] PR #51 merged: chore: bump go-imap to v2.0.0-beta.9",
			text:    "PR #51 has been merged into main by maeregzewdu.\n\nTitle: chore: bump go-imap to v2.0.0-beta.9\nMerged by: maeregzewdu\nReview: 2 approvals\n\nView: https://github.com/myapp/backend/pull/51",
			agoMin: 30, read: true,
		},
		{
			folder: "INBOX", from: "billing@digitalocean.com",
			subject: "Your DigitalOcean invoice for January 2024",
			html: `<div style="font-family:sans-serif;padding:24px;max-width:560px">
<div style="display:flex;gap:8px;align-items:center;margin-bottom:20px">
<div style="width:28px;height:28px;background:#0080FF;border-radius:50%"></div>
<strong>DigitalOcean</strong>
</div>
<h2>Invoice #INV-2024-001</h2>
<table style="width:100%;border-collapse:collapse;font-size:14px">
<tr style="border-bottom:1px solid #e5e7eb"><td style="padding:8px">Droplet (4GB RAM)</td><td style="text-align:right;padding:8px">$24.00</td></tr>
<tr style="border-bottom:1px solid #e5e7eb"><td style="padding:8px">Managed PostgreSQL (1 node)</td><td style="text-align:right;padding:8px">$15.00</td></tr>
<tr style="border-bottom:1px solid #e5e7eb"><td style="padding:8px">Spaces Object Storage</td><td style="text-align:right;padding:8px">$5.00</td></tr>
<tr><td style="padding:8px;font-weight:600">Total</td><td style="text-align:right;padding:8px;font-weight:600">$44.00</td></tr>
</table>
</div>`,
			agoMin: 120,
		},
		{
			folder: "INBOX", from: "no-reply@linear.app",
			subject: "ENG-248: 5 issues are overdue",
			html: `<div style="font-family:sans-serif;padding:24px;max-width:560px">
<h2 style="color:#6366f1">5 overdue issues need your attention</h2>
<ul style="padding-left:20px;line-height:1.8">
<li>ENG-201: Update API rate limits documentation</li>
<li>ENG-215: Write migration guide for v2</li>
<li>ENG-218: Add E2E tests for auth flow</li>
<li>ENG-223: Optimize slow GraphQL query</li>
<li>ENG-231: Fix timezone bug in scheduler</li>
</ul>
<a href="#" style="color:#6366f1;text-decoration:none;font-weight:500">View in Linear →</a>
</div>`,
			agoMin: 200, starred: true,
		},
		{
			folder: "INBOX", from: "deploy@vercel.com",
			subject: "❌ Deployment failed: myapp-staging",
			text:    "Your deployment has failed.\n\nProject: myapp-staging\nBranch: feat/new-dashboard\nCommit: e9a21bc\nError: Build failed (exit code 1)\n\nError output:\nError: Type 'string | undefined' is not assignable to type 'string'.\n  Type 'undefined' is not assignable to type 'string'.\n\nView logs: https://vercel.com/dashboard/deployments/dpl_abc123",
			agoMin:  280, starred: true,
		},
		{
			folder: "INBOX", from: "support@algolia.com",
			subject: "Your Algolia plan is expiring in 7 days",
			html: `<div style="font-family:sans-serif;padding:24px;max-width:560px">
<div style="background:#fef3c7;border:1px solid #fbbf24;border-radius:8px;padding:16px;margin-bottom:20px">
<p style="margin:0;font-weight:600;color:#92400e">Your trial ends in 7 days</p>
</div>
<p>Your Algolia Starter plan expires on <strong>January 22, 2024</strong>.</p>
<p>Upgrade to continue using search with no interruption.</p>
<a href="#" style="display:inline-block;background:#003dff;color:#fff;padding:10px 20px;border-radius:6px;text-decoration:none;font-weight:500">Upgrade Plan</a>
</div>`,
			agoMin: 400, read: true,
		},
		{
			folder: "INBOX", from: "weekly@changelog.com",
			subject: "The Changelog Weekly #571 — Go 1.22 Released",
			text:    "This week in developer news:\n\n1. Go 1.22 released with range-over-integers and improved routing\n2. Bun v1.1 brings 3.5x faster SQLite reads\n3. OpenAI launches new embedding models\n4. React 19 enters release candidate phase\n\nRead more: https://changelog.com/weekly/571",
			agoMin: 600, read: true,
		},
		{
			folder: "INBOX", from: "noreply@sentry.io",
			subject: "[CRITICAL] Unhandled exception in production",
			html: `<div style="font-family:monospace;padding:16px;background:#1a1a1a;color:#e5e5e5;border-radius:8px;max-width:600px">
<div style="color:#ef4444;font-size:16px;font-weight:600;margin-bottom:12px">🔴 CRITICAL: runtime error</div>
<div style="color:#fbbf24">runtime: goroutine stack exceeds 1000000000-byte limit</div>
<div style="color:#94a3b8;margin-top:8px;font-size:12px">
goroutine 1 [running]:<br>
main.syncLoop(0xc000118000)<br>
&nbsp;&nbsp;/app/internal/mail/sync.go:42 +0x1a8<br>
main.main()<br>
&nbsp;&nbsp;/app/cmd/catchy/main.go:95 +0x3f4
</div>
</div>`,
			agoMin: 800, starred: true,
		},
		{
			folder: "INBOX", from: "hello@resend.com",
			subject: "Your first email was sent successfully!",
			html: `<div style="font-family:sans-serif;padding:32px;max-width:560px;margin:0 auto">
<h2>🎉 First email sent!</h2>
<p>Congrats! You just sent your first email via Resend.</p>
<div style="background:#f0fdf4;border:1px solid #86efac;border-radius:8px;padding:16px;margin:16px 0">
<p style="margin:0;font-size:14px;color:#166534"><strong>From:</strong> onboarding@resend.dev<br><strong>To:</strong> staging@myapp.dev<br><strong>Subject:</strong> Hello World</p>
</div>
<p style="color:#6b7280;font-size:14px">Next steps: add a custom domain, set up webhooks, or explore the API.</p>
</div>`,
			agoMin: 1200, read: true,
		},
		// Sent folder
		{
			folder: "Sent", from: a.Email,
			subject: "Re: Q4 roadmap discussion",
			text:    "Hi team,\n\nHere's my input on the Q4 priorities:\n\n1. Complete the IMAP sync reliability work — we've had too many missed messages\n2. Add full-text search (already scoped in Phase 6)\n3. Performance improvements to the message list (virtual scrolling)\n\nLet me know if you have questions.\n\nBest,\nMaereg",
			agoMin: 500, read: true,
		},
		{
			folder: "Sent", from: a.Email,
			subject: "API credentials for staging environment",
			text:    "Hi Alex,\n\nHere are the staging API keys you asked for.\n\nPlease note:\n- These are staging-only credentials\n- Rotate them after your testing is done\n- Never commit to git!\n\nLet me know if you need anything else.\n\nThanks,\nMaereg",
			agoMin: 900, read: true,
		},
		{
			folder: "Sent", from: a.Email,
			subject: "Bug report: attachment download fails on iOS Safari",
			text:    "Hi support team,\n\nI'm seeing a consistent failure when trying to download attachments on iOS Safari 17.\n\nSteps to reproduce:\n1. Open any email with an attachment\n2. Tap the Download button\n3. Browser opens a new tab that immediately closes\n\nExpected: File download dialog appears\nActual: Nothing happens, no error shown\n\nHappens on iPhone 15 Pro, iOS 17.2, Safari 17.\n\nThanks,\nMaereg",
			agoMin: 1500, read: true,
		},
	}

	for _, sm := range messages {
		receivedAt := now.Add(-time.Duration(sm.agoMin) * time.Minute)
		sentAt := receivedAt
		id := uuid.New().String()
		accountID := a.ID
		source := "imap"
		if sm.folder == "Sent" {
			source = "sent"
		}
		raw := buildRawMIME(sm.from, a.Email, sm.subject, sm.text, sm.html, sentAt)
		msg := &model.Message{
			ID:         id,
			AccountID:  &accountID,
			Source:     source,
			Folder:     sm.folder,
			MessageID:  "<" + uuid.New().String() + "@seed.catchy>",
			Subject:    sm.subject,
			FromAddr:   sm.from,
			ToAddrs:    []string{a.Email},
			BodyText:   sm.text,
			BodyHTML:   sm.html,
			IsRead:     sm.read,
			IsStarred:  sm.starred,
			RawMIME:    raw,
			SentAt:     &sentAt,
			ReceivedAt: &receivedAt,
			CreatedAt:  receivedAt,
		}
		if err := db.CreateMessage(msg, nil); err != nil {
			slog.Warn("seed imap message", "subject", sm.subject, "err", err)
		}
	}
	return nil
}

// ── RFC 2822 message builder ───────────────────────────────────────────────────

func buildRawMIME(from, to, subject, text, html string, date time.Time) []byte {
	msgID := uuid.New().String()
	var b []byte
	b = append(b, fmt.Sprintf("From: %s\r\n", from)...)
	b = append(b, fmt.Sprintf("To: %s\r\n", to)...)
	b = append(b, fmt.Sprintf("Subject: %s\r\n", subject)...)
	b = append(b, fmt.Sprintf("Date: %s\r\n", date.Format("Mon, 02 Jan 2006 15:04:05 -0700"))...)
	b = append(b, fmt.Sprintf("Message-ID: <%s@seed.catchy>\r\n", msgID)...)
	b = append(b, "MIME-Version: 1.0\r\n"...)

	if html != "" && text != "" {
		boundary := "boundary_" + msgID[:8]
		b = append(b, fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary)...)
		b = append(b, "\r\n"...)
		b = append(b, fmt.Sprintf("--%s\r\n", boundary)...)
		b = append(b, "Content-Type: text/plain; charset=utf-8\r\n\r\n"...)
		b = append(b, []byte(text)...)
		b = append(b, []byte("\r\n")...)
		b = append(b, fmt.Sprintf("--%s\r\n", boundary)...)
		b = append(b, "Content-Type: text/html; charset=utf-8\r\n\r\n"...)
		b = append(b, []byte(html)...)
		b = append(b, []byte("\r\n")...)
		b = append(b, fmt.Sprintf("--%s--\r\n", boundary)...)
	} else if html != "" {
		b = append(b, "Content-Type: text/html; charset=utf-8\r\n\r\n"...)
		b = append(b, []byte(html)...)
	} else {
		b = append(b, "Content-Type: text/plain; charset=utf-8\r\n\r\n"...)
		if text != "" {
			b = append(b, []byte(text)...)
		}
	}
	return b
}
