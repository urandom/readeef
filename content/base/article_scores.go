package base

import (
	"errors"
	"fmt"

	"github.com/urandom/readeef/content/info"
)

type ArticleScores struct {
	Error
	RepoRelated

	info info.ArticleScores
}

func (asc ArticleScores) String() string {
	return fmt.Sprintf("Scores for article '%d'", asc.info.ArticleId)
}

func (asc *ArticleScores) Info(in ...info.ArticleScores) info.ArticleScores {
	if asc.HasErr() {
		return info.ArticleScores{}
	}

	if len(in) > 0 {
		asc.info = in[0]
	}

	return asc.info
}

func (asc ArticleScores) Validate() error {
	if asc.info.ArticleId == 0 {
		return ValidationError{errors.New("Article scores has no article id")}
	}

	return nil
}

func (asc ArticleScores) Calculate() int64 {
	return asc.info.Score1 + int64(0.1*float64(asc.info.Score2)) + int64(0.01*float64(asc.info.Score3)) + int64(0.001*float64(asc.info.Score4)) + int64(0.0001*float64(asc.info.Score5))
}
