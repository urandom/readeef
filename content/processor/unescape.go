package processor

import (
	"html"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/log"
	"github.com/urandom/readeef/parser"
)

type Unescape struct {
	log log.Log
}

func NewUnescape(l log.Log) Unescape {
	return Unescape{log: l}
}

func (p Unescape) ProcessArticles(articles []content.Article) []content.Article {
	if len(articles) == 0 {
		return articles
	}

	p.log.Infof("Unescaping articles of feed %d", articles[0].FeedID)

	for i := range articles {
		articles[i].Description = p.processDescription(articles[i].Description)
	}

	return articles
}

func (p Unescape) ProcessFeed(f parser.Feed) parser.Feed {
	for i := range f.Articles {
		f.Articles[i].Description = p.processDescription(f.Articles[i].Description)
	}

	return f
}

func (p Unescape) processDescription(description string) string {
	return html.UnescapeString(description)
}
