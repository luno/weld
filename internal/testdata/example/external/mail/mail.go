// Package mail provides a client interface for a 3rd party email service.
package mail

type Mailer struct {
}

func (m *Mailer) Send(to, from, text string) error {
	return nil
}

type option func()

func New(opts ...option) (*Mailer, error) {
	return new(Mailer), nil
}
