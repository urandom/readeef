package base

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/advancedlogic/GoOse"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/webfw/renderer"
	"github.com/urandom/webfw/util"
)

type ReadabilityExtractor struct {
	key string
}

type GooseExtractor struct {
	renderer renderer.Renderer
}

type readability struct {
	Content   string
	Title     string
	LeadImage string `json:"lead_image_url"`
}

func NewReadabilityExtractor(key string) ReadabilityExtractor {
	return ReadabilityExtractor{key: key}
}

func NewGooseExtractor(templateDir string) GooseExtractor {
	renderer := renderer.NewRenderer(templateDir, "raw.tmpl")
	renderer.Delims("{%", "%}")

	return GooseExtractor{renderer: renderer}
}

func (e ReadabilityExtractor) Extract(link string) (data data.ArticleExtract, err error) {
	url := fmt.Sprintf("http://readability.com/api/content/v1/parser?url=%s&token=%s",
		url.QueryEscape(link), e.key,
	)

	var r readability
	var resp *http.Response

	resp, err = http.Get(url)

	if err != nil {
		return
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)

	err = dec.Decode(&r)
	if err != nil {
		err = fmt.Errorf("Error extracting content from %s: %v", link, err)
		return
	}

	data.Title = r.Title
	data.Content = r.Content
	data.TopImage = r.LeadImage
	data.Language = "en"
	return
}

func (e GooseExtractor) Extract(link string) (data data.ArticleExtract, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
		}
	}()

	g := goose.New()
	/* TODO: preserve links */
	formatted := g.ExtractFromUrl(link)

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
