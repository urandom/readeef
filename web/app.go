package web

import (
	"html/template"
	"net/http"

	"github.com/alexedwards/scs/session"
	"github.com/pkg/errors"
	"github.com/urandom/handler/lang"
	"github.com/urandom/handler/method"
)

const (
	baseTmpl = "templates/base.tmpl"
	appTmpl  = "templates/app.tmpl"
)

type mainPayload struct {
	Language string
}

func MainHandler(fs http.FileSystem) (http.Handler, error) {
	tmpl, err := prepareTemplate(template.New("main"), fs, baseTmpl, appTmpl)
	if err != nil {
		return nil, errors.Wrap(err, "preparing main template")
	}

	return method.HTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error

		data := lang.Data(r)
		if err = tmpl.Funcs(requestFuncMaps(r)).Execute(w, mainPayload{
			Language: data.Current.String(),
		}); err == nil {
			if err = session.PutBool(r, visitorKey, true); err == nil {
				err = session.Save(w, r)
			}
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Header().Set("X-Readeef", "1")
	}), method.GET, method.HEAD), nil
}
