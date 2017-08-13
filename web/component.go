package web

import (
	"html/template"
	"net/http"
	"path"

	"github.com/urandom/handler/lang"
)

const (
	componentsPrefix = "/templates/components/"
	rawTmpl          = "templates/raw.tmpl"
)

type componentPayload struct {
	APIPattern   string
	Language     string
	Languages    []string
	HasLanguages bool
}

type cacheMap map[string]*template.Template
type cacheResult struct {
	tmpl *template.Template
	err  error
}

func ComponentHandler(fs http.FileSystem) (http.HandlerFunc, error) {
	ops := make(chan func(cacheMap))
	go func() {
		cache := cacheMap{}

		for op := range ops {
			op(cache)
		}
	}()

	getTmpl := func(name string) cacheResult {
		res := make(chan cacheResult, 1)

		ops <- func(cache cacheMap) {
			tmpl := cache[name]
			if tmpl == nil {
				tmpl, err := prepareTemplate(template.New("components"), fs, rawTmpl, componentsPrefix+name+".tmpl")
				cache[name] = tmpl

				res <- cacheResult{tmpl, err}
				return
			}

			res <- cacheResult{tmpl, nil}
		}

		return <-res
	}

	return func(w http.ResponseWriter, r *http.Request) {
		name := path.Base(r.URL.Path)
		res := getTmpl(name)
		if res.err != nil {
			http.Error(w, res.err.Error(), http.StatusInternalServerError)
			return
		}

		data := lang.Data(r)
		languages := make([]string, len(data.Languages))
		for i := range data.Languages {
			languages[i] = data.Languages[i].String()
		}
		err := res.tmpl.Funcs(requestFuncMaps(r)).Execute(w, componentPayload{
			APIPattern:   "/api/v2",
			Language:     data.Current.String(),
			Languages:    languages,
			HasLanguages: len(data.Languages) > 0,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}, nil
}
