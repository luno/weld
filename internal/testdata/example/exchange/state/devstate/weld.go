// +build weld

package devstate

import (
	"example/backends/providers"
	exchange_ops "example/exchange/ops"
	"example/exchange/state"

	"github.com/luno/weld"
)

//go:generate weld

var _ = weld.NewSpec(
	weld.NewSet(state.ChanProvider, providers.WeldDev),
	weld.Existing(new(exchange_ops.Backends)),
)
