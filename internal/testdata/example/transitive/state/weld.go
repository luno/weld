//go:build weld

package state

import (
	"example/backends/providers"
	"example/transitive/ops"

	"github.com/luno/weld"
)

//go:generate weld -tags=!dev
// Note that github.com/luno/weld/internal/gen_test.go generates this specific state_gen.go as well.

var _ = weld.NewSpec(
	weld.NewSet(
		providers.WeldProd,
		ops.NewFoo,
		ops.NewBar,
		ops.NewBaz,
		ops.NewQux,
	),
	weld.GenUnion(new(ops.Backends)),
)
