package logical

import (
	"example/identity/email"
	"example/identity/email/ops"
	"example/identity/users/db"
)

// Backends defines the dependencies required for serving API requests.
// Note it is a subset of the main users ops Backends.
type Backends interface {
	UsersDB() *db.UsersDB
	Email() email.Client
}

// adapter wraps Backends to satisfy root users ops Backends.
type adapter struct {
	Backends
	extend
}

// extend allows the above adapter to embed both Backends interfaces.
type extend struct {
	ops.Backends
}
