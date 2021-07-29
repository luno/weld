package state

import (
	"example/exchange"

	"github.com/luno/weld"
)

var ChanProvider = weld.NewSet(NewModelChan)

func NewModelChan() chan<- exchange.Model {
	return make(chan<- exchange.Model)
}
