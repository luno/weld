package users

import "context"

// Client defines the root users service interface.
type Client interface {
	Ping(context.Context) error

	// TODO(rig): Add more interface methods.
}

type ClientProvider interface {
	UsersClient() Client
}
