package extract

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	goOse "github.com/advancedlogic/GoOse"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
)

const (
	rawTmpl   = "templates/raw.tmpl"
	gooseTmpl = "templates/goose-format-result.tmpl"
)

type goose struct {
	template *template.Template
	buf      bytes.Buffer
}

func WithGoose(templateDir string, fs http.FileSystem) (Generator, error) {
	tmpl, err := prepareTemplate(template.New("goose").Delims("{%", "%}"), fs, rawTmpl, gooseTmpl)
	if err != nil {
		return nil, errors.Wrap(err, "parsing goose template")
	}

	return goose{template: tmpl}, nil
}

func (e goose) Generate(link string) (extract content.Extract, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
		}
	}()

	g := goOse.New()
	/* TODO: preserve links */
	formatted, err := g.ExtractFromURL(link)

	content := formatted.CleanedText
	e.buf.Reset()

	paragraphs := strings.Split(content, "\n")
	var html []template.HTML

	for _, p := range paragraphs {
		if strings.TrimSpace(p) != "" {
			html = append(html, template.HTML(p))
		}
	}

	if err = e.template.Execute(&e.buf, map[string]interface{}{
		"paragraphs": html, "topImage": formatted.TopImage,
	}); err != nil {
		return extract, errors.Wrap(err, "executing goose template")
	}

	extract.Content = e.buf.String()
	extract.Title = formatted.Title
	extract.TopImage = formatted.TopImage
	extract.Language = formatted.MetaLang

	return extract, err
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
