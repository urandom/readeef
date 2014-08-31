package web

import (
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/middleware"
	"github.com/urandom/webfw/renderer"
)

func RegisterControllers(dispatcher *webfw.Dispatcher, apiPattern string) {
	dispatcher.Handle(NewApp())
	dispatcher.Handle(NewComponent(dispatcher.Config.Renderer.Dir, apiPattern))

	dispatcher.Renderer = renderer.NewRenderer(dispatcher.Config.Renderer.Dir,
		dispatcher.Config.Renderer.Base)

	dispatcher.Renderer.Delims("{%", "%}")

	middleware.InitializeDefault(dispatcher)
}
