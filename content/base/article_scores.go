package base

import (
	"errors"
	"fmt"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

type ArticleScores struct {
	Error
	RepoRelated

	data data.ArticleScores
}

func (asc ArticleScores) String() string {
	return fmt.Sprintf("Scores for article '%d'", asc.data.ArticleId)
}

func (asc *ArticleScores) Data(d ...data.ArticleScores) data.ArticleScores {
	if asc.HasErr() {
		return data.ArticleScores{}
	}

	if len(d) > 0 {
		asc.data = d[0]
	}

	return asc.data
}

func (asc ArticleScores) Validate() error {
	if asc.data.ArticleId == 0 {
		return content.NewValidationError(errors.New("Article scores has no article id"))
	}

	return nil
}

func (asc ArticleScores) Calculate() int64 {
	return asc.data.Score1 + int64(0.1*float64(asc.data.Score2)) + int64(0.01*float64(asc.data.Score3)) + int64(0.001*float64(asc.data.Score4)) + int64(0.0001*float64(asc.data.Score5))
}
