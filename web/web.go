package web

import (
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/middleware"
)

func RegisterControllers(dispatcher *webfw.Dispatcher) {
	dispatcher.Handle(NewApp())
	dispatcher.Handle(NewComponent(dispatcher.Config.Renderer.Dir))

	middleware.InitializeDefault(dispatcher)
}
