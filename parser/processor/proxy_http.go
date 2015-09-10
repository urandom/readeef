package processor

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/urandom/readeef/parser"
	"github.com/urandom/webfw"
)

type ProxyHTTP struct {
	logger webfw.Logger
}

func NewProxyHTTP(l webfw.Logger) ProxyHTTP {
	return ProxyHTTP{logger: l}
}

func (p ProxyHTTP) Process(f parser.Feed) parser.Feed {
	p.logger.Infof("Proxying urls of feed '%s'\n", f.Title)

	for i := range f.Articles {
		if d, err := goquery.NewDocumentFromReader(strings.NewReader(f.Articles[i].Description)); err == nil {
			if proxyArticleLinks(d) {
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

func proxyArticleLinks(d *goquery.Document) bool {
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

		s.SetAttr("src", "/proxy?url="+url.QueryEscape(u.String()))

		changed = true
		return
	})

	return changed
}
