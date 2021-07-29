// Package mail provides another client interface for a 3rd party email service.
// It clashes with example/external/mail;
package mail

type MailerLegacy struct {
}

func (m *MailerLegacy) Send(to, from, text string) error {
	return nil
}

func New() (*MailerLegacy, error) {
	return new(MailerLegacy), nil
}
