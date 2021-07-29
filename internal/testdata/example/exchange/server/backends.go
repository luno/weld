package server

import (
	"example/exchange/db"
	"example/identity/email"
	"example/identity/users"
)

type Backends interface {
	ExchangeDB() *db.ExchangeDB
	Email() email.Client
	Users() users.Client
}
