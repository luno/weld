package devstate

// Code generated by weld. DO NOT EDIT.

import (
	"example/exchange"
	exchange_db "example/exchange/db"
	exchange_ops "example/exchange/ops"
	exchange_state "example/exchange/state"
	"example/external/versioned"
	versioned_v1 "example/external/versioned/v1"
	"example/identity/email"
	email_client_dev "example/identity/email/client/dev"
	email_client_logical "example/identity/email/client/logical"
	email_db "example/identity/email/db"
	"example/identity/users"
	users_client_dev "example/identity/users/client/dev"
	users_client_logical "example/identity/users/client/logical"
	users_db "example/identity/users/db"

	"github.com/luno/jettison/errors"
)

func MakeBackends() (exchange_ops.Backends, error) {
	var (
		b   backendsImpl
		err error
	)

	b.email, err = email_client_dev.Make(&b)
	if err != nil {
		return nil, errors.Wrap(err, "email client dev make")
	}

	b.emailDB, err = email_db.Connect()
	if err != nil {
		return nil, errors.Wrap(err, "email db connect")
	}

	b.exchangeDB, err = exchange_db.Connect()
	if err != nil {
		return nil, errors.Wrap(err, "exchange db connect")
	}

	b.genericStringFunc = exchange_state.NewGenericStringType()

	b.modelChan = exchange_state.NewModelChan()

	b.users, err = users_client_dev.Make(&b)
	if err != nil {
		return nil, errors.Wrap(err, "users client dev make")
	}

	b.usersDB, err = users_db.Connect()
	if err != nil {
		return nil, errors.Wrap(err, "users db connect")
	}

	b.versioned = versioned.New()

	return &b, nil
}

type backendsImpl struct {
	email             email.Client
	emailDB           *email_db.EmailDB
	exchangeDB        *exchange_db.ExchangeDB
	genericStringFunc exchange_ops.GenericStringType
	modelChan         chan<- exchange.Model
	users             users.Client
	usersDB           *users_db.UsersDB
	versioned         *versioned_v1.Service
}

func (b *backendsImpl) Email() email.Client {
	return b.email
}

func (b *backendsImpl) EmailDB() *email_db.EmailDB {
	return b.emailDB
}

func (b *backendsImpl) ExchangeDB() *exchange_db.ExchangeDB {
	return b.exchangeDB
}

func (b *backendsImpl) GenericStringFunc() exchange_ops.GenericStringType {
	return b.genericStringFunc
}

func (b *backendsImpl) ModelChan() chan<- exchange.Model {
	return b.modelChan
}

func (b *backendsImpl) Users() users.Client {
	return b.users
}

func (b *backendsImpl) UsersDB() *users_db.UsersDB {
	return b.usersDB
}

func (b *backendsImpl) Versioned() *versioned_v1.Service {
	return b.versioned
}

// Transitive dependency interface assertions.
var (
	_ email_client_logical.Backends = (*backendsImpl)(nil)
	_ users_client_logical.Backends = (*backendsImpl)(nil)
)
