package eventable

import (
	"encoding/json"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

const (
	FeedUpdateEvent = "feed-update"
	FeedDeleteEvent = "feed-delete"
)

type FeedUpdateData struct {
	Feed        content.Feed
	NewArticles []content.Article
}

func (f FeedUpdateData) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{}

	data["feedID"] = f.Feed.ID

	ids := make([]content.ArticleID, len(f.NewArticles))
	for i := range f.NewArticles {
		ids[i] = f.NewArticles[i].ID
	}
	data["articleIDs"] = ids

	return json.Marshal(data)
}

func (f FeedUpdateData) FeedID() content.FeedID {
	return f.Feed.ID
}

type FeedDeleteData struct {
	Feed content.Feed
}

func (f FeedDeleteData) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{}

	data["feedID"] = f.Feed.ID

	return json.Marshal(data)
}

func (f FeedDeleteData) FeedID() content.FeedID {
	return f.Feed.ID
}

type feedRepo struct {
	repo.Feed
	eventBus bus
	log      log.Log
}

func (r feedRepo) Update(feed *content.Feed) ([]content.Article, error) {
	articles, err := r.Feed.Update(feed)

	if err == nil && len(articles) > 0 {
		r.log.Debugf("Dispatching feed update event")

		r.eventBus.Dispatch(
			FeedUpdateEvent,
			FeedUpdateData{*feed, articles},
		)

		r.log.Debugf("Dispatch of feed update event end")
	}

	return articles, err
}

func (r feedRepo) Delete(feed content.Feed) error {
	err := r.Feed.Delete(feed)

	if err == nil {
		r.log.Debugf("Dispatching feed delete event")

		r.eventBus.Dispatch(
			FeedDeleteEvent,
			FeedDeleteData{feed},
		)

		r.log.Debugf("Dispatch of feed delete event end")
	}

	return err
}
