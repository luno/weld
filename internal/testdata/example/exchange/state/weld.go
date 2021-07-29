// +build weld

package state

import (
	"example/backends/providers"
	exchange_ops "example/exchange/ops"

	"github.com/luno/weld"
)

//go:generate weld -tags=!dev
// Note that github.com/luno/weld/internal/gen_test.go generates this specific state_gen.go as well.

var _ = weld.NewSpec(
	weld.NewSet(ChanProvider, providers.WeldProd),
	weld.Existing(new(exchange_ops.Backends)),
)
