package processor

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"

	"github.com/PuerkitoBio/goquery"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/log"
	"github.com/urandom/readeef/parser"
)

type RelativeURL struct {
	log log.Log
}

func NewRelativeURL(log log.Log) RelativeURL {
	return RelativeURL{log: log}
}

func (p RelativeURL) ProcessArticles(articles []content.Article) []content.Article {
	if len(articles) == 0 {
		return articles
	}

	p.log.Infof("Proxying urls of feed '%d'\n", articles[0].FeedID)

	for i := range articles {
		articles[i].Description = p.processDescription(articles[i].Description)
	}

	return articles
}

func (p RelativeURL) ProcessFeed(f parser.Feed) parser.Feed {
	p.log.Infof("Converting urls of feed '%s' to protocol-relative schemes\n", f.Title)

	for i := range f.Articles {
		f.Articles[i].Description = p.processDescription(f.Articles[i].Description)
	}

	return f
}

func (p RelativeURL) processDescription(description string) string {
	if d, err := goquery.NewDocumentFromReader(strings.NewReader(description)); err == nil {
		if relativizeArticleLinks(d) {
			if content, err := d.Html(); err == nil {
				// net/http tries to provide valid html, adding html, head and body tags
				content = content[strings.Index(content, "<body>")+6 : strings.LastIndex(content, "</body>")]

				return content
			}
		}
	}

	return description
}

func relativizeArticleLinks(d *goquery.Document) bool {
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
