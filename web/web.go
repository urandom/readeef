package web

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"time"

	"golang.org/x/text/language"

	ttemplate "text/template"

	"github.com/alexedwards/scs/session"
	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/pkg/errors"
	"github.com/urandom/handler/lang"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/config"
	"github.com/urandom/readeef/log"
)

const (
	visitorKey = "visitor"
)

type sessionWrapper struct{}

func Mux(fs http.FileSystem, engine session.Engine, config config.Config, log log.Log) (http.Handler, error) {
	languages, err := readeef.GetLanguages(fs)
	if err != nil {
		return nil, errors.WithMessage(err, "getting supported languages")
	}

	mux := http.NewServeMux()
	mainHandlers, err := mainRoutes(fs, engine, languages, log)
	if err != nil {
		return nil, err
	}

	if hasProxy(config) {
		mux.HandleFunc("/proxy", ProxyHandler)
	}

	fileServer := http.FileServer(fs)
	dir, err := fs.Open("/static")
	if err != nil {
		return nil, errors.Wrap(err, "opening /static dir")
	}
	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, errors.Wrap(err, "reading /static dir")
	}

	rootNameSet := map[string]struct{}{}
	for _, f := range files {
		rootNameSet[f.Name()] = struct{}{}
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/web/", http.StatusFound)
			return
		}

		cleaned := path.Clean(r.URL.Path)
		if cleaned[0] == '/' {
			cleaned = cleaned[1:]
		}
		base := cleaned
		idx := strings.Index(base, "/")
		if idx != -1 {
			base = base[:idx]
		}
		if _, ok := rootNameSet[base]; ok {
			r.URL.Path = path.Join("/static", r.URL.Path)
			fileServer.ServeHTTP(w, r)
			return
		}

		mainHandlers.ServeHTTP(w, r)
	})

	session := session.Manage(engine, session.Lifetime(240*time.Hour))
	return session(mux), nil
}

func mainRoutes(fs http.FileSystem, engine session.Engine, languages []language.Tag, log log.Log) (http.Handler, error) {
	mux := http.NewServeMux()

	componentHandler, err := ComponentHandler(fs)
	if err != nil {
		return nil, errors.Wrap(err, "creating component handler")
	}

	mux.Handle("/component/", componentHandler)

	main, err := MainHandler(fs)
	if err != nil {
		return nil, errors.WithMessage(err, "creating main handler")
	}

	mux.Handle("/web/", main)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/web/", http.StatusFound)
	})

	return wrapInI18N(mux, sessionWrapper{}, languages, log), nil
}

func wrapInI18N(next http.Handler, s sessionWrapper, languages []language.Tag, log log.Log) http.Handler {
	if len(languages) > 0 {
		return lang.I18N(next, lang.Languages(languages), lang.Session(s), lang.Logger(log))
	}

	return next
}

func hasProxy(config config.Config) bool {
	hasProxy := false
	for _, p := range config.FeedParser.Processors {
		if p == "proxy-http" {
			hasProxy = true
			break
		}
	}

	if !hasProxy {
		for _, p := range config.Content.Article.Processors {
			if p == "proxy-http" {
				hasProxy = true
				break
			}
		}
	}

	return hasProxy
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
		"url": func(url string, prefix ...string) string {
			var p string
			if len(prefix) > 0 {
				p = prefix[0]
			}
			return lang.URL(url, p, langData)
		},
	}
}

func prepareTemplate(t *template.Template, fs http.FileSystem, paths ...string) (*template.Template, error) {
	t = t.Funcs(template.FuncMap{
		"__": func(message string, data ...interface{}) (template.HTML, error) {
			return template.HTML(message), nil
		},
		"url": func(url string, prefix ...string) string {
			return url
		},
	}).Delims("{%", "%}")
	for _, path := range paths {
		f, err := fs.Open(path)
		if err != nil {
			return nil, errors.Wrapf(err, "opening template %s", path)
		}

		t, err = parseTemplate(t, f)
		if err != nil {
			f.Close()
			return nil, errors.WithMessage(err, fmt.Sprintf("parsing template %s", path))
		}

		if err = f.Close(); err != nil {
			return nil, errors.Wrapf(err, "closing template %s", path)
		}
	}

	return t, nil
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
