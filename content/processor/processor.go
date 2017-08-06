package processor

import "github.com/urandom/readeef/content"

type Article interface {
	Process([]content.Article) []content.Article
}

type Articles []Article

func (processors Articles) Process(articles []content.Article) []content.Article {
	for _, p := range processors {
		articles = p.Process(articles)
	}

	return articles
}
