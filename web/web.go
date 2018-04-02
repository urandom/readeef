package web

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"time"

	"golang.org/x/text/language"

	ttemplate "text/template"

	"github.com/alexedwards/scs"
	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/pkg/errors"
	"github.com/urandom/handler/lang"
	"github.com/urandom/readeef/config"
	"github.com/urandom/readeef/log"
)

const (
	visitorKey = "visitor"
)

type e struct{}

func Mux(fs http.FileSystem, sessionManager *scs.Manager, config config.Config, log log.Log) (http.Handler, error) {
	mux := http.NewServeMux()

	if hasProxy(config) {
		mux.Handle(
			"/proxy",
			http.TimeoutHandler(http.HandlerFunc(ProxyHandler(sessionManager)), 10*time.Second, ""),
		)
	}

	fileServer := http.FileServer(fs)

	mux.Handle("/", http.TimeoutHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := sessionManager.Load(r)
		err := session.PutBool(w, visitorKey, true)

		if err != nil {
			log.Printf("Error saving session: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		current := path.Join("/", config.UI.Path, r.URL.Path)
		path, err := closestParent(current, fs)
		if err != nil {
			log.Printf("Error getting file path for %s: %+v", r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		r.URL.Path = path

		// Firefox does not seem to be able to load its own cached files.
		r.Header.Del("If-Modified-Since")

		fileServer.ServeHTTP(w, r)
	}), time.Second, ""))

	return sessionManager.Use(mux), nil
}

func closestParent(p string, fs http.FileSystem) (string, error) {
	f, err := fs.Open(p)
	if err == nil {
		stat, err := f.Stat()
		f.Close()
		if err != nil {
			return "", errors.Wrapf(err, "getting file stat for %s", p)
		}

		if stat.IsDir() {
			return p + "/", nil
		}
		return p, nil
	}

	return closestParent(path.Dir(p), fs)
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
