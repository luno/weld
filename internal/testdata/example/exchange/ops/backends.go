package ops

import (
	"example/exchange"
	"example/exchange/db"
	versioned "example/external/versioned/v1"
	"example/identity/email"
	"example/identity/users"
)

type TestFunc[T any, C any] func(a T, c C) string

type GenericStringType = TestFunc[exchange.Model, string]

type Backends interface {
	GenericStringFunc() GenericStringType
	ExchangeDB() *db.ExchangeDB
	Email() email.Client
	Users() users.Client
	ModelChan() chan<- exchange.Model
	Versioned() *versioned.Service
}
