package web

import (
	"html/template"
	"net/http"
	"path"

	"golang.org/x/text/language"

	"github.com/urandom/handler/lang"
)

const (
	componentsPrefix = "/templates/components/"
	rawTmpl          = "templates/raw.tmpl"
)

type componentPayload struct {
	APIPattern   string
	Language     string
	Languages    []language.Tag
	HasLanguages bool
}

func ComponentHandler(fs http.FileSystem) (http.HandlerFunc, error) {
	cache := map[string]*template.Template{}

	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		name := path.Base(r.URL.Path)

		tmpl := cache[name]
		if tmpl == nil {
			tmpl, err = prepareTemplate(template.New("components"), fs, rawTmpl, componentsPrefix+name+".tmpl")
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
	}, nil
}
