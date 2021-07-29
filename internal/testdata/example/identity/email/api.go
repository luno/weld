package email

import "context"

// Client defines the root email service interface.
type Client interface {
	Ping(context.Context) error

	// TODO(rig): Add more interface methods.
}

type ClientProvider interface {
	EmailClient() Client
}
