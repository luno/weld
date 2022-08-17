//go:build weld

package devstate

import (
	"example/backends/providers"
	"example/identity/state"

	"github.com/luno/weld"
)

//go:generate weld
// Note that github.com/luno/weld/internal/gen_test.go generates this specific state_gen.go as well.

var _ = weld.NewSpec(
	weld.NewSet(providers.WeldDev),
	weld.Existing(new(state.Backends)),
)
