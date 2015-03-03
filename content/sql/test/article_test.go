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
