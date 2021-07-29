package exchange

import "context"

// Client defines the root exchange service interface.
type Client interface {
	Ping(context.Context) error

	// TODO(rig): Add more interface methods.
}

type ClientProvider interface {
	ExchangeClient() Client
}
