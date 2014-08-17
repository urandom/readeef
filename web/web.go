package web

import (
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/middleware"
)

func RegisterControllers(dispatcher *webfw.Dispatcher) {
	dispatcher.Handle(NewApp())

	middleware.InitializeDefault(dispatcher)
}
