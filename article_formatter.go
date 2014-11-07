package readeef

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/advancedlogic/GoOse"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/renderer"
	"github.com/urandom/webfw/util"
)

type ArticleFormatting struct {
	Content  string
	Title    string
	TopImage string
	Language string
}

type readability struct {
	Content   string
	Title     string
	LeadImage string `json:"lead_image_url"`
}

var (
	gooseRenderer       renderer.Renderer
	rendererInitialized bool
)

func ArticleFormatter(webfwConfig webfw.Config, readeefConfig Config, a Article) (ArticleFormatting, error) {
	var f ArticleFormatting
	var err error

	if readeefConfig.ArticleFormatter.ReadabilityKey != "" {

		url := fmt.Sprintf("http://readability.com/api/content/v1/parser?url=%s&token=%s",
			url.QueryEscape(a.Link), readeefConfig.ArticleFormatter.ReadabilityKey,
		)

		var r readability
		var resp *http.Response

		resp, err = http.Get(url)

		if err != nil {
			return f, err
		}

		defer resp.Body.Close()
		dec := json.NewDecoder(resp.Body)

		err = dec.Decode(&r)
		if err != nil {
			return f, err
		}

		f.Title = r.Title
		f.Content = r.Content
		f.TopImage = r.LeadImage
		f.Language = "en"

		return f, nil
	}

	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
		}
	}()

	g := goose.New()
	/* TODO: preserve links */
	formatted := g.ExtractFromUrl(a.Link)

	content := formatted.CleanedText
	buf := util.BufferPool.GetBuffer()
	defer util.BufferPool.Put(buf)

	initRenderer(webfwConfig)

	paragraphs := strings.Split(content, "\n")
	var html []template.HTML

	for _, p := range paragraphs {
		if strings.TrimSpace(p) != "" {
			html = append(html, template.HTML(p))
		}
	}

	gooseRenderer.Render(buf,
		renderer.RenderData{"paragraphs": html, "topImage": formatted.TopImage},
		nil, "goose-format-result.tmpl")

	f.Content = buf.String()
	f.Title = formatted.Title
	f.TopImage = formatted.TopImage
	f.Language = formatted.MetaLang

	return f, nil
}

func initRenderer(webfwConfig webfw.Config) {
	if !rendererInitialized {
		gooseRenderer = renderer.NewRenderer(webfwConfig.Renderer.Dir, "raw.tmpl")
		gooseRenderer.Delims("{%", "%}")
	}
}
