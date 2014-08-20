package web

import (
	"net/http"

	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
)

type App struct {
	webfw.BaseController
}

func NewApp() App {
	return App{
		webfw.NewBaseController("/", webfw.MethodAll, ""),
	}
}

func (con App) Handler(c context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rnd := webfw.GetRenderer(c)
		rnd.Delims("{%", "%}")
		err := rnd.Render(w, nil, c.GetAll(r), "app.tmpl")
		if err != nil {
			webfw.GetLogger(c).Print(err)
		}
	}
}
