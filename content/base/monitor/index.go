package monitor

import (
	"fmt"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/webfw"
)

type Index struct {
	provider content.SearchProvider
	logger   webfw.Logger
}

func NewIndex(sp content.SearchProvider, l webfw.Logger) Index {
	return Index{provider: sp, logger: l}
}

func (i Index) FeedUpdated(feed content.Feed) error {
	i.logger.Infof("Updating article search index for feed '%s'\n", feed)

	return i.provider.BatchIndex(feed.NewArticles(), data.BatchAdd)
}

func (i Index) FeedDeleted(feed content.Feed) error {
	i.logger.Infof("Deleting article search index for feed '%s'\n", feed)

	articles := feed.AllArticles()

	if feed.HasErr() {
		return fmt.Errorf("Error deleting all articles of %s from the search index: %v\n", feed, feed.Err())
	} else {
		i.logger.Infof("Deleting article search index for feed '%s'\n", feed)

		return i.provider.BatchIndex(articles, data.BatchDelete)
	}

}
