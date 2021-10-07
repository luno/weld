// +build weld

package state

import (
	"example/samevar/ops1"
	"example/samevar/ops2"
	"example/samevar/pool"

	"github.com/luno/weld"
)

//go:generate weld

var _ = weld.NewSpec(
	weld.NewSet(
		pool.NewFooPool,
		pool.NewBarPool,
	),
	weld.GenUnion(
		new(ops1.Backends),
		new(ops2.Backends),
	),
)
