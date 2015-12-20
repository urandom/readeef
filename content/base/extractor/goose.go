package extractor

import (
	"errors"
	"fmt"
	"html/template"
	"strings"

	goose "github.com/advancedlogic/GoOse"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/webfw/fs"
	"github.com/urandom/webfw/renderer"
	"github.com/urandom/webfw/util"
)

type Goose struct {
	renderer renderer.Renderer
}

func NewGoose(templateDir string) (content.Extractor, error) {
	rawTmpl := "raw.tmpl"
	f, err := fs.DefaultFS.OpenRoot(templateDir, rawTmpl)
	if err != nil {
		return nil, fmt.Errorf("Goose extractor requires %s template in %s: %v\n", rawTmpl, templateDir, err)
	}
	f.Close()
	renderer := renderer.NewRenderer(templateDir, rawTmpl)
	renderer.Delims("{%", "%}")

	return Goose{renderer: renderer}, nil
}

func (e Goose) Extract(link string) (data data.ArticleExtract, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
		}
	}()

	g := goose.New()
	/* TODO: preserve links */
	formatted, err := g.ExtractFromURL(link)

	content := formatted.CleanedText
	buf := util.BufferPool.GetBuffer()
	defer util.BufferPool.Put(buf)

	paragraphs := strings.Split(content, "\n")
	var html []template.HTML

	for _, p := range paragraphs {
		if strings.TrimSpace(p) != "" {
			html = append(html, template.HTML(p))
		}
	}

	e.renderer.Render(buf,
		renderer.RenderData{"paragraphs": html, "topImage": formatted.TopImage},
		nil, "goose-format-result.tmpl")

	data.Content = buf.String()
	data.Title = formatted.Title
	data.TopImage = formatted.TopImage
	data.Language = formatted.MetaLang
	return
}
