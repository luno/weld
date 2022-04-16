//go:build weld
// +build weld

package state

import (
	backends_providers "example/backends/providers"
	"example/sort_with_bind"
	"example/sort_with_bind/ops"
	"example/sort_with_bind/providers"

	"github.com/luno/weld"
)

//go:generate weld

var _ = weld.NewSpec(
	weld.NewSet(
		providers.NewFoo,
		weld.NewSet(providers.NewBar, weld.Bind(new(sort_with_bind.Bar), new(*providers.Bar))),
		providers.NewBaz,
		backends_providers.WeldProd,
	),
	weld.GenUnion(new(ops.Backends)),
)
