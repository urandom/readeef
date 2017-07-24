package monitor

import (
	"fmt"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

type Index struct {
	provider content.SearchProvider
	log      readeef.Logger
}

func NewIndex(sp content.SearchProvider, l readeef.Logger) Index {
	return Index{provider: sp, log: l}
}

func (i Index) FeedUpdated(feed content.Feed) error {
	i.log.Infof("Updating article search index for feed '%s'\n", feed)

	return i.provider.BatchIndex(feed.NewArticles(), data.BatchAdd)
}

func (i Index) FeedDeleted(feed content.Feed) error {
	i.log.Infof("Deleting article search index for feed '%s'\n", feed)

	articles := feed.AllArticles()

	if feed.HasErr() {
		return fmt.Errorf("Error deleting all articles of %s from the search index: %v\n", feed, feed.Err())
	} else {
		i.log.Infof("Deleting article search index for feed '%s'\n", feed)

		return i.provider.BatchIndex(articles, data.BatchDelete)
	}

}
