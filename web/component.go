package web

import (
	"net/http"

	"github.com/urandom/readeef"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/middleware"
	"github.com/urandom/webfw/renderer"
)

type Component struct {
	webfw.BaseController
	dir        string
	apiPattern string
}

func NewComponent(dir, apiPattern string) Component {
	return Component{
		BaseController: webfw.NewBaseController("/component/:name", webfw.MethodAll, ""),
		dir:            dir,
		apiPattern:     apiPattern,
	}
}

func (con Component) Handler(c context.Context) http.HandlerFunc {
	dispatcher := webfw.GetDispatcher(c)
	mw, i18nFound := dispatcher.Middleware("I18N")

	rnd := renderer.NewRenderer(con.dir, "raw.tmpl")
	rnd.Delims("{%", "%}")

	if i18nFound {
		if i18n, ok := mw.(middleware.I18N); ok {
			rnd.Funcs(i18n.TemplateFuncMap())
		}
	} else {
		readeef.Debug.Println("I18N middleware not found")
	}

	return func(w http.ResponseWriter, r *http.Request) {
		params := webfw.GetParams(c, r)

		err := rnd.Render(w, renderer.RenderData{"apiPattern": con.apiPattern},
			c.GetAll(r), "components/"+params["name"]+".tmpl")

		if err != nil {
			webfw.GetLogger(c).Print(err)
		}
	}
}
