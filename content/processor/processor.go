package processor

import "github.com/urandom/readeef/content"

type Article interface {
	Process([]content.Article) []content.Article
}

type Articles []Article

func (processors Articles) Process(articles []Article) []Article {
	for _, p := range processors {
		articles = p.Process(articles)
	}

	return articles
}
