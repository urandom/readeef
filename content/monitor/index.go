package monitor

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/content/search"
	"github.com/urandom/readeef/log"
)

type Index struct {
	repo     repo.Article
	provider search.Provider
	log      log.Log
}

func NewIndex(repo repo.Article, sp search.Provider, l log.Log) Index {
	return Index{repo: repo, provider: sp, log: l}
}

func (i Index) FeedUpdated(feed content.Feed, articles []content.Article) error {
	i.log.Infof("Updating article search index for feed %s", feed)

	return i.provider.BatchIndex(articles, search.BatchAdd)
}

func (i Index) FeedDeleted(feed content.Feed) error {
	i.log.Infof("Deleting article search index for feed %s", feed)

	articles, err := i.repo.All(content.FeedIDs([]content.FeedID{feed.ID}))

	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("getting feed %s articles", feed))
	}
	i.log.Infof("Deleting article search index for feed %s", feed)

	return i.provider.BatchIndex(articles, search.BatchDelete)

}
