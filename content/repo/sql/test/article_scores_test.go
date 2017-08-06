package test

import (
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

func createArticleScores(d data.ArticleScores) (asc content.ArticleScores) {
	asc = repo.ArticleScores()
	asc.Data(d)

	asc.Update()

	return
}
