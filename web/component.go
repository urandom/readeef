package web

import (
	"html/template"
	"net/http"
	"path"

	"golang.org/x/text/language"

	"github.com/pkg/errors"
	"github.com/urandom/handler/lang"
	"github.com/urandom/handler/method"
)

const (
	componentsPrefix = "/templates/components/"
)

type componentPayload struct {
	APIPattern   string
	Language     string
	Languages    []language.Tag
	HasLanguages bool
}

func ComponentHandler(fs http.FileSystem) (http.Handler, error) {
	baseTmpl, err := prepareTemplate(template.New("components"), fs)
	if err != nil {
		return nil, errors.Wrap(err, "preparing components template")
	}

	cache := map[string]*template.Template{}

	return method.HTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error

		name := path.Base(r.URL.Path)

		tmpl := cache[name]
		if tmpl == nil {
			tmpl, err = prepareTemplate(baseTmpl, fs, componentsPrefix+name+".tmpl")
			cache[name] = tmpl
		}

		if err == nil {
			data := lang.Data(r)
			err = tmpl.Funcs(requestFuncMaps(r)).Execute(w, componentPayload{
				APIPattern:   "/api/v2",
				Language:     data.Current.String(),
				Languages:    data.Languages,
				HasLanguages: len(data.Languages) > 0,
			})
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}), method.GET, method.HEAD), nil
}
