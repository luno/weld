//go:build weld

package state

import (
	"example/variadic/ops"

	"github.com/luno/weld"
)

//go:generate weld -tags=!dev
// Note that github.com/luno/weld/internal/gen_test.go generates this specific state_gen.go as well.

var _ = weld.NewSpec(
	weld.NewSet(
		ops.NewFooForTesting,
	),
	weld.GenUnion(new(ops.Backends)),
)
