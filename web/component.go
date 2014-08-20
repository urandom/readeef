package web

import (
	"net/http"

	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/renderer"
)

type Component struct {
	webfw.BaseController
	dir string
}

func NewComponent(dir string) Component {
	return Component{
		BaseController: webfw.NewBaseController("/component/:name", webfw.MethodAll, ""),
		dir:            dir,
	}
}

func (con Component) Handler(c context.Context) http.HandlerFunc {
	rnd := renderer.NewRenderer(con.dir, "raw.tmpl")
	rnd.Delims("{%", "%}")

	return func(w http.ResponseWriter, r *http.Request) {
		params := webfw.GetParams(c, r)

		err := rnd.Render(w, nil, c.GetAll(r), "components/"+params["name"]+".tmpl")
		if err != nil {
			webfw.GetLogger(c).Print(err)
		}
	}
}
