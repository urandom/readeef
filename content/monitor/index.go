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
			go processIndexUpdateEvent(data, provider, log)
		case eventable.FeedDeleteData:
			go processIndexDeleteEvent(data, provider, log)
		}
	}
}

func processIndexUpdateEvent(data eventable.FeedUpdateData, provider search.Provider, log log.Log) {
	log.Infof("Updating article search index for feed %s", data.Feed)

	if err := provider.BatchIndex(data.NewArticles, search.BatchAdd); err != nil {
		log.Printf("Error adding articles from %s to search index: %+v", data.Feed, err)
	}
}

func processIndexDeleteEvent(data eventable.FeedDeleteData, provider search.Provider, log log.Log) {
	log.Infof("Deleting article search index for feed %s", data.Feed)

	if err := provider.RemoveFeed(data.Feed.ID); err != nil {
		log.Printf("Error removing feed %s from search index: %+v", data.Feed, err)
	}
}
