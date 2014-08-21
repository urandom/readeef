package web

import (
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/middleware"
)

func RegisterControllers(dispatcher *webfw.Dispatcher, apiPattern string) {
	dispatcher.Handle(NewApp())
	dispatcher.Handle(NewComponent(dispatcher.Config.Renderer.Dir, apiPattern))

	middleware.InitializeDefault(dispatcher)
}
