// + build weld

package state

import (
	"example/backends/providers"
	email_logical "example/identity/email/client/logical"
	email_ops "example/identity/email/ops"
	users_logical "example/identity/users/client/logical"
	users_ops "example/identity/users/ops"

	"github.com/luno/weld"
)

//go:generate weld -verbose -tags=!dev
// Note that github.com/luno/weld/internal/gen_test.go generates this specific state_gen.go as well.

var _ = weld.NewSpec(
	weld.NewSet(email_logical.Provider, users_logical.Provider, providers.WeldProd),
	weld.GenUnion(new(email_ops.Backends), new(users_ops.Backends)),
)
