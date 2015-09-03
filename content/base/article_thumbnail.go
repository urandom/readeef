package base

import (
	"errors"
	"fmt"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

type ArticleThumbnail struct {
	Error
	RepoRelated

	data data.ArticleThumbnail
}

func (at ArticleThumbnail) String() string {
	return fmt.Sprintf("Thumbnail for article '%d'", at.data.ArticleId)
}

func (at *ArticleThumbnail) Data(d ...data.ArticleThumbnail) data.ArticleThumbnail {
	if at.HasErr() {
		return data.ArticleThumbnail{}
	}

	if len(d) > 0 {
		at.data = d[0]
	}

	return at.data
}

func (at ArticleThumbnail) Validate() error {
	if at.data.ArticleId == 0 {
		return content.NewValidationError(errors.New("Article thumbnail has no article id"))
	}

	return nil
}
