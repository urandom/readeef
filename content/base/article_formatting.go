package base

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/advancedlogic/GoOse"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/webfw/renderer"
	"github.com/urandom/webfw/util"
)

type readability struct {
	Content   string
	Title     string
	LeadImage string `json:"lead_image_url"`
}

var (
	gooseRenderer       renderer.Renderer
	rendererInitialized bool
	mutex               = &sync.Mutex{}
)

func (a *Article) Format(templateDir, readabilityKey string) (f data.ArticleFormatting) {
	if a.HasErr() {
		return
	}

	if readabilityKey != "" {
		url := fmt.Sprintf("http://readability.com/api/content/v1/parser?url=%s&token=%s",
			url.QueryEscape(a.Data().Link), readabilityKey,
		)

		var r readability
		var resp *http.Response

		resp, err := http.Get(url)

		if err != nil {
			a.Err(err)
			return
		}

		defer resp.Body.Close()
		dec := json.NewDecoder(resp.Body)

		err = dec.Decode(&r)
		if err != nil {
			a.Err(fmt.Errorf("Error formatting article %s: %v", a, err))
			return
		}

		f.Title = r.Title
		f.Content = r.Content
		f.TopImage = r.LeadImage
		f.Language = "en"

		return
	}

	defer func() {
		if r := recover(); r != nil {
			a.Err(errors.New(fmt.Sprintf("%v", r)))
		}
	}()

	g := goose.New()
	/* TODO: preserve links */
	formatted := g.ExtractFromUrl(a.Data().Link)

	content := formatted.CleanedText
	buf := util.BufferPool.GetBuffer()
	defer util.BufferPool.Put(buf)

	initRenderer(templateDir)

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

	return
}

func initRenderer(templateDir string) {
	mutex.Lock()
	if !rendererInitialized {
		gooseRenderer = renderer.NewRenderer(templateDir, "raw.tmpl")
		gooseRenderer.Delims("{%", "%}")
	}
	mutex.Unlock()
}
