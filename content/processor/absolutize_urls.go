package processor

import (
	"net/url"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/urandom/readeef/log"
	"github.com/urandom/readeef/parser"
)

type AbsolutizeURLs struct {
	log log.Log
}

func NewAbsolutizeURLs(l log.Log) AbsolutizeURLs {
	return AbsolutizeURLs{log: l}
}

func (p AbsolutizeURLs) ProcessFeed(f parser.Feed) parser.Feed {
	p.log.Infof("Converting relative urls of feed '%s' to absolute\n", f.Title)

	for i := range f.Articles {
		if d, err := goquery.NewDocumentFromReader(strings.NewReader(f.Articles[i].Description)); err == nil {
			articleLink, err := url.Parse(f.Articles[i].Link)
			if err != nil {
				continue
			}

			if convertRelativeLinksToAbsolute(d, articleLink) {
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

func convertRelativeLinksToAbsolute(d *goquery.Document, articleLink *url.URL) bool {
	changed := false
	d.Find("[src]").Each(func(i int, s *goquery.Selection) {
		val, ok := s.Attr("src")
		if !ok {
			return
		}

		u, err := url.Parse(val)
		if err != nil {
			return
		}

		if !u.IsAbs() {
			u.Scheme = articleLink.Scheme

			if u.Host == "" {
				u.Host = articleLink.Host

				if u.Path[0] != '/' {
					u.Path = path.Join(path.Dir(articleLink.Path), u.Path)
				}
			}

		}

		s.SetAttr("src", u.String())

		changed = true
		return
	})

	return changed
}
