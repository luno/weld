// Package providers defines weld provider sets for the
// the core repo dependency injection.
package providers

import (
	exchange_dev "example/exchange/client/dev"
	exchange_grpc "example/exchange/client/grpc"
	exchange_db "example/exchange/db"
	"example/external/mail"
	mail_legacy "example/external/mail/mail"
	"example/external/versioned"
	email_dev "example/identity/email/client/dev"
	email_grpc "example/identity/email/client/grpc"
	email_db "example/identity/email/db"
	users_dev "example/identity/users/client/dev"
	users_grpc "example/identity/users/client/grpc"
	users_db "example/identity/users/db"

	"github.com/luno/weld"
)

var (
	// WeldProd defines the production (and staging) weld provider sets.
	// Note the order of weld providers set are important.
	WeldProd = weld.NewSet(
		GRPC,
		DB,
		External,
	)

	// WeldDev defines the dev weld provider sets.
	// Note the order of weld providers set are important.
	WeldDev = weld.NewSet(
		Dev,
		DB,
		// We could add DevExternal stubs here before actual external clients.
		External,
	)

	// GRPC provides grpc clients.
	GRPC = weld.NewSet(
		users_grpc.Provider,
		email_grpc.Provider,
		exchange_grpc.Provider,
	)

	// DB provides db connections.
	DB = weld.NewSet(
		email_db.Connect,
		users_db.Connect,
		exchange_db.Connect,
	)

	// External provides external 3rd party clients.
	DevExternal = weld.NewSet()

	// External provides external 3rd party clients.
	External = weld.NewSet(
		mail.New,
		mail_legacy.New,
		versioned.New,
	)

	// Dev provides dev clients (either grpc or logical or stub).
	// Note these require transitive dependencies.
	Dev = weld.NewSet(
		exchange_dev.Make,
		users_dev.Make,
		email_dev.Make,
	)
)
