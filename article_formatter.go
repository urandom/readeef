package readeef

import (
	"encoding/json"
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

type Readability struct {
	Content   string
	LeadImage string `json:"lead_image_url"`
}

var (
	gooseRenderer       renderer.Renderer
	rendererInitialized bool
)

func ArticleFormatter(webfwConfig webfw.Config, readeefConfig Config, a Article) (string, string, error) {
	if readeefConfig.ArticleFormatter.ReadabilityKey != "" {

		url := fmt.Sprintf("http://readability.com/api/content/v1/parser?url=%s&token=%s",
			url.QueryEscape(a.Link), readeefConfig.ArticleFormatter.ReadabilityKey,
		)

		var r Readability

		response, err := http.Get(url)

		if err != nil {
			return "", "", err
		}

		defer response.Body.Close()
		dec := json.NewDecoder(response.Body)

		err = dec.Decode(&r)
		if err != nil {
			return "", "", err
		}

		return r.Content, r.LeadImage, nil
	}

	defer func() {
		if r := recover(); r != nil {
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

	return buf.String(), formatted.TopImage, nil
}

func initRenderer(webfwConfig webfw.Config) {
	if !rendererInitialized {
		gooseRenderer = renderer.NewRenderer(webfwConfig.Renderer.Dir, "raw.tmpl")
		gooseRenderer.Delims("{%", "%}")
	}
}
