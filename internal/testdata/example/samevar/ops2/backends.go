package ops2

import "example/samevar/pool"

type Backends interface {
	GetPool() *pool.BarPool
}
