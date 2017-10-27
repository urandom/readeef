package logging

import (
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type tagRepo struct {
	repo.Tag

	log log.Log
}

func (r tagRepo) Get(id content.TagID, user content.User) (content.Tag, error) {
	start := time.Now()

	tag, err := r.Tag.Get(id, user)

	r.log.Infof("repo.Tag.Get took %s", time.Now().Sub(start))

	return tag, err
}

func (r tagRepo) ForUser(user content.User) ([]content.Tag, error) {
	start := time.Now()

	tags, err := r.Tag.ForUser(user)

	r.log.Infof("repo.Tag.ForUser took %s", time.Now().Sub(start))

	return tags, err
}

func (r tagRepo) ForFeed(feed content.Feed, user content.User) ([]content.Tag, error) {
	start := time.Now()

	tags, err := r.Tag.ForFeed(feed, user)

	r.log.Infof("repo.Tag.ForFeed took %s", time.Now().Sub(start))

	return tags, err
}

func (r tagRepo) FeedIDs(tag content.Tag, user content.User) ([]content.FeedID, error) {
	start := time.Now()

	ids, err := r.Tag.FeedIDs(tag, user)

	r.log.Infof("repo.Tag.FeedIDs took %s", time.Now().Sub(start))

	return ids, err
}
