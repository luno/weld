package ops

import (
	"example/exchange"
	"example/identity/email"
	"example/identity/users/db"
)

// Backends is the main email Backends required for API requests and background loops.
type Backends interface {
	// API requests and background loops
	UsersDB() *db.UsersDB
	Email() email.Client

	// background loops only
	Exchange() exchange.Client
}
