package processor

import (
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/parser"
)

type Article interface {
	ProcessArticles([]content.Article) []content.Article
}

type Feed interface {
	ProcessFeed(parser.Feed) parser.Feed
}

type Articles []Article

func (processors Articles) Process(articles []content.Article) []content.Article {
	for _, p := range processors {
		articles = p.ProcessArticles(articles)
	}

	return articles
}
