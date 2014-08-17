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
		err := webfw.GetRenderCtx(c, r)(w, nil, "app.tmpl")
		if err != nil {
			webfw.GetLogger(c).Print(err)
		}
	}
}
