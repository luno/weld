package server

import (
	"example/identity/email"
	"example/identity/users/db"
	"example/identity/users/ops"
)

// Backends defines the dependencies required for serving API requests.
// Note it is a subset of the main users ops Backends.
type Backends interface {
	UsersDB() *db.UsersDB
	Email() email.Client
}

// adapter wraps Backends and extends it to satisfy main email ops Backends.
type adapter struct {
	Backends
	extend
}

// extend allows the above adapter to embed both Backends interfaces.
type extend struct {
	ops.Backends
}
