package eventable

import (
	"encoding/json"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

const (
	FeedUpdateEvent  = "feed-update"
	FeedDeleteEvent  = "feed-delete"
	FeedSetTagsEvent = "feed-set-tags"
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

type FeedSetTagsData struct {
	Feed content.Feed
	User content.User
	Tags []*content.Tag
}

func (f FeedSetTagsData) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{}

	data["feedID"] = f.Feed.ID
	data["user"] = f.User.Login

	tagIDs := make([]content.TagID, len(f.Tags))
	for i := range f.Tags {
		tagIDs[i] = f.Tags[i].ID
	}

	data["tagIDs"] = tagIDs

	return json.Marshal(data)
}

func (f FeedSetTagsData) FeedID() content.FeedID {
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

func (r feedRepo) SetUserTags(feed content.Feed, user content.User, tags []*content.Tag) error {
	err := r.Feed.SetUserTags(feed, user, tags)

	if err == nil {
		r.log.Debugf("Dispatching feed set tags event")

		r.eventBus.Dispatch(
			FeedSetTagsEvent,
			FeedSetTagsData{feed, user, tags},
		)

		r.log.Debugf("Dispatch of feed set tags end")
	}

	return err
}
