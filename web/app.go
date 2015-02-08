package web

import (
	"net/http"

	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/renderer"
)

type App struct {
	webfw.BasePatternController
}

func NewApp() App {
	return App{}
}

func (con App) Patterns() []webfw.MethodIdentifierTuple {
	return []webfw.MethodIdentifierTuple{
		webfw.MethodIdentifierTuple{"/", webfw.MethodAll, ""},
		webfw.MethodIdentifierTuple{"/web/*history", webfw.MethodAll, "history"},
	}
}

func (con App) Handler(c context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rnd := webfw.GetRenderer(c)
		action := webfw.GetMultiPatternIdentifier(c, r)

		data := renderer.RenderData{}
		if action == "history" {
			params := webfw.GetParams(c, r)
			data["history"] = "/web/" + params["history"]
		}

		err := rnd.Render(w, data, c.GetAll(r), "app.tmpl")
		if err != nil {
			webfw.GetLogger(c).Print(err)
		}
	})
}
