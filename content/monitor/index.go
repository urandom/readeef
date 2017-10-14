package monitor

import (
	"github.com/urandom/readeef/content/repo/eventable"
	"github.com/urandom/readeef/content/search"
	"github.com/urandom/readeef/log"
)

func Index(service eventable.Service, provider search.Provider, log log.Log) {
	for event := range service.Listener() {
		switch data := event.Data.(type) {
		case eventable.FeedUpdateData:
			log.Infof("Updating article search index for feed %s", data.Feed)

			if err := provider.BatchIndex(data.NewArticles, search.BatchAdd); err != nil {
				log.Printf("Error adding articles from %s to search index: %+v", data.Feed, err)
			}
		case eventable.FeedDeleteData:
			log.Infof("Deleting article search index for feed %s", data.Feed)

			if err := provider.RemoveFeed(data.Feed.ID); err != nil {
				log.Printf("Error removing feed %s from search index: %+v", data.Feed, err)
			}
		}
	}
}
