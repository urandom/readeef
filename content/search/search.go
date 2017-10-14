package search

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
)

type indexOperation int

const (
	BatchAdd indexOperation = iota + 1
	BatchDelete
)

type Provider interface {
	IsNewIndex() bool
	Search(string, content.User, ...content.QueryOpt) ([]content.Article, error)
	BatchIndex(articles []content.Article, op indexOperation) error
	RemoveFeed(content.FeedID) error
}

func Reindex(p Provider, repo repo.Article) error {
	limit := 2000
	offset := 0

	for {
		articles, err := repo.All(content.Paging(limit, offset))
		if err != nil {
			return errors.WithMessage(err, fmt.Sprintf(
				"getting articles in window %d-%d", offset, offset+limit,
			))
		}

		if err = p.BatchIndex(articles, BatchAdd); err != nil {
			return errors.WithMessage(err, "adding batch to index")
		}

		if len(articles) < limit {
			return nil
		}

		offset += limit

	}
}
