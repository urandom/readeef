package monitor

import (
	"fmt"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/content/search"
	"github.com/urandom/readeef/log"
)

type Index struct {
	provider search.Provider
	log      log.Log
}

func NewIndex(sp search.Provider, l log.Log) Index {
	return Index{provider: sp, log: l}
}

func (i Index) FeedUpdated(feed content.Feed) error {
	i.log.Infof("Updating article search index for feed %s", feed)

	return i.provider.BatchIndex(feed.NewArticles(), data.BatchAdd)
}

func (i Index) FeedDeleted(feed content.Feed) error {
	i.log.Infof("Deleting article search index for feed %s", feed)

	articles := feed.AllArticles()

	if feed.HasErr() {
		return fmt.Errorf("Error deleting all articles of %s from the search index: %v", feed, feed.Err())
	}
	i.log.Infof("Deleting article search index for feed %s", feed)

	return i.provider.BatchIndex(articles, data.BatchDelete)

}
