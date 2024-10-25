package state

import (
	"fmt"

	"github.com/luno/weld"
	"github.com/luno/weld/internal/testdata/example/exchange"
	"github.com/luno/weld/internal/testdata/example/exchange/ops"
)

var ChanProvider = weld.NewSet(NewModelChan)

func NewModelChan() chan<- exchange.Model {
	return make(chan<- exchange.Model)
}

func NewGenericStringType() ops.GenericStringType {
	return func(a exchange.Model, b string) string {
		return fmt.Sprintf("%s%s", a, b)
	}
}
