package ops

import (
	"example/exchange"
	"example/exchange/db"
	versioned "example/external/versioned/v1"
	"example/identity/email"
	"example/identity/users"
)

type Backends interface {
	ExchangeDB() *db.ExchangeDB
	Email() email.Client
	Users() users.Client
	ModelChan() chan<- exchange.Model
	Versioned() *versioned.Service
}
