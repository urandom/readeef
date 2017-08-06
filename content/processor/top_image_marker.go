package processor

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/urandom/readeef/log"
	"github.com/urandom/readeef/parser"
	"golang.org/x/net/html"
)

type TopImageMarker struct {
	log log.Log
}

func NewTopImageMarker(l log.Log) TopImageMarker {
	return TopImageMarker{log: l}
}

func (p TopImageMarker) ProcessFeed(f parser.Feed) parser.Feed {
	p.log.Infof("Locating suitable top images in articles of '%s'\n", f.Title)

	for i := range f.Articles {
		if d, err := goquery.NewDocumentFromReader(strings.NewReader(f.Articles[i].Description)); err == nil {
			if markTopImage(d) {
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

func markTopImage(d *goquery.Document) bool {
	changed := false

	totalTextCount := len(d.Text())
	img := d.Find("img")

	if img.Length() == 0 {
		return changed
	}

	if totalTextCount == 0 {
		img.AddClass("top-image")
		changed = true
	} else {
		afterTextCount := len(img.NextAll().Text())
		// Add any text-node siblings of the image
		for n := img.Get(0).NextSibling; n != nil; n = n.NextSibling {
			if n.Type == html.TextNode {
				afterTextCount += len(n.Data)
			}
		}

		img.Parents().Each(func(i int, s *goquery.Selection) {
			afterTextCount += len(s.NextAll().Text())
		})

		if totalTextCount-afterTextCount <= afterTextCount {
			img.AddClass("top-image")
			changed = true
		}
	}

	return changed
}
