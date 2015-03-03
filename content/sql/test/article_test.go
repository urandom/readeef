package test

import (
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

func createArticle(d data.Article) (a content.Article) {
	a = repo.Article()
	a.Data(d)

	return
}

func createUserArticle(u content.User, d data.Article) (ua content.Article) {
	ua = repo.UserArticle(u)
	ua.Data(d)

	return
}

func createScoredArticle(u content.User, d data.Article) (sa content.Article) {
	sa = repo.ScoredArticle()
	sa.Data(d)

	return sa
}
