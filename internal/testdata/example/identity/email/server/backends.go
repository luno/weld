package server

import (
	"example/identity/email/db"
	"example/identity/email/ops"
)

// Backends defines the dependencies required for serving API requests.
// Note it is a subset of the root main ops Backends.
type Backends interface {
	EmailDB() *db.EmailDB
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
