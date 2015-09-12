package processor

import (
	"net/url"
	"strings"
	"text/template"

	"github.com/PuerkitoBio/goquery"
	"github.com/urandom/readeef/parser"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/util"
)

type ProxyHTTP struct {
	logger      webfw.Logger
	urlTemplate *template.Template
}

func NewProxyHTTP(l webfw.Logger, urlTemplate string) (ProxyHTTP, error) {
	t, err := template.New("proxy-http-url-template").Parse(urlTemplate)
	if err != nil {
		return ProxyHTTP{}, err
	}

	return ProxyHTTP{logger: l, urlTemplate: t}, nil
}

func (p ProxyHTTP) Process(f parser.Feed) parser.Feed {
	p.logger.Infof("Proxying urls of feed '%s'\n", f.Title)

	for i := range f.Articles {
		if d, err := goquery.NewDocumentFromReader(strings.NewReader(f.Articles[i].Description)); err == nil {
			if proxyArticleLinks(d, p.urlTemplate) {
				if content, err := d.Html(); err == nil {
					// net/http tries to provide valid html, adding html, head and body tags
					content = content[strings.Index(content, "<body>")+6 : strings.LastIndex(content, "</body>")]

					f.Articles[i].Description = content
				}
			}
		}
	}

	return f
}

func proxyArticleLinks(d *goquery.Document, urlTemplate *template.Template) bool {
	changed := false
	d.Find("[src]").Each(func(i int, s *goquery.Selection) {
		var val string
		var ok bool

		if val, ok = s.Attr("src"); !ok {
			return
		}

		u, err := url.Parse(val)
		if err != nil || !u.IsAbs() || u.Scheme != "http" {
			return
		}

		buf := util.BufferPool.GetBuffer()
		defer util.BufferPool.Put(buf)

		if err := urlTemplate.Execute(buf, url.QueryEscape(u.String())); err != nil {
			return
		}

		s.SetAttr("src", buf.String())

		changed = true
		return
	})

	return changed
}
