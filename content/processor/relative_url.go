package processor

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/log"
	"github.com/urandom/readeef/parser/processor"
)

type RelativeURL struct {
	log log.Log
}

func NewRelativeURL(log log.Log) RelativeURL {
	return RelativeURL{log: log}
}

func (p RelativeURL) Process(articles []content.Article) []content.Article {
	if len(articles) == 0 {
		return articles
	}

	p.log.Infof("Proxying urls of feed '%d'\n", articles[0].FeedID)

	for i := range articles {
		if d, err := goquery.NewDocumentFromReader(strings.NewReader(articles[i].Description)); err == nil {
			if processor.RelativizeArticleLinks(d) {
				if content, err := d.Html(); err == nil {
					// net/http tries to provide valid html, adding html, head and body tags
					content = content[strings.Index(content, "<body>")+6 : strings.LastIndex(content, "</body>")]

					articles[i].Description = content
				}
			}
		}
	}

	return articles
}
