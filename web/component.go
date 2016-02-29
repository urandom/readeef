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
	webfw.BasePatternController
	dispatcher *webfw.Dispatcher
	apiPattern string
}

func NewComponent(dispatcher *webfw.Dispatcher, apiPattern string) Component {
	return Component{
		BasePatternController: webfw.NewBasePatternController("/component/:name", webfw.MethodAll, ""),
		dispatcher:            dispatcher,
		apiPattern:            apiPattern,
	}
}

func (con Component) Handler(c context.Context) http.Handler {
	i18nmw, i18nFound := con.dispatcher.Middleware("I18N")
	urlmw, urlFound := con.dispatcher.Middleware("Url")
	logger := webfw.GetLogger(c)
	cfg := readeef.GetConfig(c)

	rnd := renderer.NewRenderer(con.dispatcher.Config.Renderer.Dir, "raw.tmpl")
	rnd.Delims("{%", "%}")

	if cfg.Logger.Level == "debug" {
		rnd.SkipCache(true)
	}

	if i18nFound {
		if i18n, ok := i18nmw.(middleware.I18N); ok {
			rnd.Funcs(i18n.TemplateFuncMap())
		}
	} else {
		logger.Infoln("I18N middleware not found")
	}

	if urlFound {
		if url, ok := urlmw.(middleware.Url); ok {
			rnd.Funcs(url.TemplateFuncMap(c))
		}
	} else {
		logger.Infoln("Url middleware not found")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := webfw.GetParams(c, r)

		if r.Method != "HEAD" {
			err := rnd.Render(w, renderer.RenderData{"apiPattern": con.apiPattern, "config": cfg},
				c.GetAll(r), "components/"+params["name"]+".tmpl")

			if err != nil {
				webfw.GetLogger(c).Print(err)
			}
		}
	})
}
