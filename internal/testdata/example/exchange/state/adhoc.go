package state

import (
	"exchange"
	"exchange/ops"
	"fmt"

	"github.com/luno/weld"
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
