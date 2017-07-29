package processor

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/parser"
	"golang.org/x/net/html"
)

type RelativeURL struct {
	log readeef.Logger
}

func NewRelativeURL(l readeef.Logger) RelativeURL {
	return RelativeURL{log: l}
}

func (p RelativeURL) Process(f parser.Feed) parser.Feed {
	p.log.Infof("Converting urls of feed '%s' to protocol-relative schemes\n", f.Title)

	for i := range f.Articles {
		if d, err := goquery.NewDocumentFromReader(strings.NewReader(f.Articles[i].Description)); err == nil {
			if RelativizeArticleLinks(d) {
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

func RelativizeArticleLinks(d *goquery.Document) bool {
	changed := false
	d.Find("[src]").Each(func(i int, s *goquery.Selection) {
		var val string
		var ok bool

		if val, ok = s.Attr("src"); !ok {
			return
		}

		u, err := url.Parse(val)
		if err != nil || !u.IsAbs() {
			return
		}

		u.Scheme = ""
		s.SetAttr("src", u.String())
		if n := s.Get(0); n.Type == html.ElementNode && n.Data == "img" {
			s.SetAttr("onerror", "this.onerror = null; this.src = '"+val+"'")
		}

		changed = true
		return
	})

	return changed
}
