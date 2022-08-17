//go:build weld

package state

import (
	"example/backends/providers"

	"github.com/luno/weld"
)

//go:generate weld -tags=!dev
// Note that github.com/luno/weld/internal/gen_test.go generates this specific state_gen.go as well.

var _ = weld.NewSpec(
	weld.NewSet(
		providers.WeldProd,
		FooProvider,
	),
	weld.GenUnion(
		new(OneBackends),
		new(TwoBackends),
	),
)
