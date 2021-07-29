package ops

import (
	"example/exchange"
	"example/external/mail"
	legacy_mail "example/external/mail/mail"
	"example/identity/email/db"
	"example/identity/users"
)

// Backends is the main email Backends required for API requests and background loops.
type Backends interface {
	// API requests and background loops
	EmailDB() *db.EmailDB

	// background loops only
	Users() users.Client
	Exchange() exchange.Client
	Mailer() *mail.Mailer
	MailerLegacy() *legacy_mail.MailerLegacy
}
