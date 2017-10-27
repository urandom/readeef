package logging

import (
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type feedRepo struct {
	repo.Feed

	log log.Log
}

func (r feedRepo) Get(id content.FeedID, user content.User) (content.Feed, error) {
	start := time.Now()

	feed, err := r.Feed.Get(id, user)

	r.log.Infof("repo.Feed.Get took %s", time.Now().Sub(start))

	return feed, err
}

func (r feedRepo) FindByLink(link string) (content.Feed, error) {
	start := time.Now()

	feed, err := r.Feed.FindByLink(link)

	r.log.Infof("repo.Feed.FindByLink took %s", time.Now().Sub(start))

	return feed, err
}

func (r feedRepo) ForUser(user content.User) ([]content.Feed, error) {
	start := time.Now()

	feeds, err := r.Feed.ForUser(user)

	r.log.Infof("repo.Feed.ForUser took %s", time.Now().Sub(start))

	return feeds, err
}

func (r feedRepo) ForTag(tag content.Tag, user content.User) ([]content.Feed, error) {
	start := time.Now()

	feeds, err := r.Feed.ForTag(tag, user)

	r.log.Infof("repo.Feed.ForTag took %s", time.Now().Sub(start))

	return feeds, err
}

func (r feedRepo) All() ([]content.Feed, error) {
	start := time.Now()

	feeds, err := r.Feed.All()

	r.log.Infof("repo.Feed.All took %s", time.Now().Sub(start))

	return feeds, err
}

func (r feedRepo) IDs() ([]content.FeedID, error) {
	start := time.Now()

	feeds, err := r.Feed.IDs()

	r.log.Infof("repo.Feed.IDs took %s", time.Now().Sub(start))

	return feeds, err
}

func (r feedRepo) Unsubscribed() ([]content.Feed, error) {
	start := time.Now()

	feeds, err := r.Feed.Unsubscribed()

	r.log.Infof("repo.Feed.Unsubscribed took %s", time.Now().Sub(start))

	return feeds, err
}

func (r feedRepo) Update(feed *content.Feed) ([]content.Article, error) {
	start := time.Now()

	articles, err := r.Feed.Update(feed)

	r.log.Infof("repo.Feed.Update took %s", time.Now().Sub(start))

	return articles, err
}

func (r feedRepo) Delete(feed content.Feed) error {
	start := time.Now()

	err := r.Feed.Delete(feed)

	r.log.Infof("repo.Feed.Delete took %s", time.Now().Sub(start))

	return err
}

func (r feedRepo) Users(feed content.Feed) ([]content.User, error) {
	start := time.Now()

	users, err := r.Feed.Users(feed)

	r.log.Infof("repo.Feed.Users took %s", time.Now().Sub(start))

	return users, err
}

func (r feedRepo) AttachTo(feed content.Feed, user content.User) error {
	start := time.Now()

	err := r.Feed.AttachTo(feed, user)

	r.log.Infof("repo.Feed.AttachTo took %s", time.Now().Sub(start))

	return err
}

func (r feedRepo) DetachFrom(feed content.Feed, user content.User) error {
	start := time.Now()

	err := r.Feed.DetachFrom(feed, user)

	r.log.Infof("repo.Feed.DetachFrom took %s", time.Now().Sub(start))

	return err
}

func (r feedRepo) SetUserTags(feed content.Feed, user content.User, tags []*content.Tag) error {
	start := time.Now()

	err := r.Feed.SetUserTags(feed, user, tags)

	r.log.Infof("repo.Feed.SetUserTags took %s", time.Now().Sub(start))

	return err
}
