package web

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/text/language"

	ttemplate "text/template"

	"github.com/alexedwards/scs/session"
	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/pkg/errors"
	"github.com/urandom/handler/lang"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/config"
)

const (
	visitorKey = "visitor"
)

type sessionWrapper struct{}

func Mux(fs http.FileSystem, engine session.Engine, config config.Config) (http.Handler, error) {
	hasProxy := false
	for _, p := range config.FeedParser.Processors {
		if p == "proxy-http" {
			hasProxy = true
			break
		}
	}

	if !hasProxy {
		for _, p := range config.Content.ArticleProcessors {
			if p == "proxy-http" {
				hasProxy = true
				break
			}
		}
	}

	mux := http.NewServeMux()

	mainHandler, err := MainHandler(fs)
	if err != nil {
		return nil, errors.Wrap(err, "creating main handler")
	}
	mux.Handle("/web/", mainHandler)

	componentHandler, err := ComponentHandler(fs)
	if err != nil {
		return nil, errors.Wrap(err, "creating component handler")
	}
	mux.Handle("/web/component/", componentHandler)

	var handler http.Handler = mux

	if hasProxy {
		mux.Handle("/web/proxy", ProxyHandler())

		manager := session.Manage(engine, session.Lifetime(10*24*time.Hour))
		handler = manager(mux)
	}

	if languages, err := readeef.GetLanguages(fs); err == nil {
		if len(languages) > 0 {
			handler = lang.I18N(
				handler,
				lang.Languages(languages),
				lang.Session(sessionWrapper{}),
			)
		}
	} else {
		return nil, errors.WithMessage(err, "getting supported languages")
	}

	return handler, nil
}

func (s sessionWrapper) Get(r *http.Request, key string) (string, error) {
	return session.GetString(r, key)
}

func (s sessionWrapper) Set(r *http.Request, key, value string) error {
	return session.PutString(r, key, value)
}

func requestFuncMaps(r *http.Request) template.FuncMap {
	langData := lang.Data(r)
	return template.FuncMap{
		"__": func(message string, data ...interface{}) (template.HTML, error) {
			if len(langData.Languages) == 0 {
				return template.HTML(message), nil
			}
			return t(message, langData.Current.String(), "en-US", data...)
		},
		"url": func(url, prefix string) string {
			return lang.URL(url, prefix, langData)
		},
	}
}

func prepareTemplate(t *template.Template, fs http.FileSystem, paths ...string) (*template.Template, error) {
	for _, path := range paths {
		f, err := fs.Open(path)
		if err != nil {
			return nil, errors.Wrapf(err, "opening template %s", path)
		}

		t, err = parseTemplate(t, f)
		if err != nil {
			f.Close()
			return nil, errors.Wrapf(err, "parsing template %s", path)
		}

		if err = f.Close(); err != nil {
			return nil, errors.Wrapf(err, "closing template %s", path)
		}
	}

	return t.Delims("{%", "%}").Funcs(template.FuncMap{
		"__": func(message string, data ...interface{}) (template.HTML, error) {
			return template.HTML(message), nil
		},
		"url": func(url, prefix string) string {
			return url
		},
	}), nil
}

func parseTemplate(t *template.Template, r io.Reader) (*template.Template, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "reading template data from reader")
	}

	if t, err = t.Parse(string(b)); err != nil {
		return nil, errors.Wrap(err, "parsing template data")
	}

	return t, nil
}

func t(message, lang, fallback string, data ...interface{}) (template.HTML, error) {
	var count interface{}
	hasCount := false

	if len(data)%2 == 1 {
		if !isNumber(data[0]) {
			return "", errors.New("The count argument must be a number")
		}
		count = data[0]
		hasCount = true

		data = data[1:]
	}

	dataMap := map[string]interface{}{}
	for i := 0; i < len(data); i += 2 {
		dataMap[data[i].(string)] = data[i+1]
	}

	T, err := i18n.Tfunc(lang, fallback)

	if err != nil {
		return "", err
	}

	var translated string
	if hasCount {
		translated = T(message, count, dataMap)
	} else {
		translated = T(message, dataMap)
	}

	if translated == message {
		// Doesn't have a translation mapping, we have to do the template evaluation by hand
		t, err := ttemplate.New("i18n").Parse(message)

		if err != nil {
			return "", err
		}

		var buf bytes.Buffer

		if err := t.Execute(&buf, dataMap); err != nil {
			return "", err
		}

		return template.HTML(buf.String()), nil
	} else {
		return template.HTML(translated), nil
	}
}

func isNumber(n interface{}) bool {
	switch n.(type) {
	case int, int8, int16, int32, int64, string:
		return true
	}
	return false
}

func languageTags(langs []string) []language.Tag {
	tags := make([]language.Tag, 0, len(langs))

	for _, l := range langs {
		tag := language.Make(l)
		if !tag.IsRoot() {
			tags = append(tags, tag)
		}
	}

	return tags
}
