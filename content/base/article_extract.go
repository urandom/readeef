package base

import (
	"errors"
	"fmt"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

type ArticleExtract struct {
	Error
	RepoRelated

	data data.ArticleExtract
}

func (ae ArticleExtract) String() string {
	return fmt.Sprintf("Extract for article '%d'", ae.data.ArticleId)
}

func (ae *ArticleExtract) Data(d ...data.ArticleExtract) data.ArticleExtract {
	if ae.HasErr() {
		return data.ArticleExtract{}
	}

	if len(d) > 0 {
		ae.data = d[0]
	}

	return ae.data
}

func (ae ArticleExtract) Validate() error {
	if ae.data.ArticleId == 0 {
		return content.NewValidationError(errors.New("Article extract has no article id"))
	}

	return nil
}
