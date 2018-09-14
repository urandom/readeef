package kv

import (
	"os"
	"path/filepath"

	"github.com/asdine/storm"
	"github.com/asdine/storm/codec/msgpack"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type Service struct {
	user         repo.User
	tag          repo.Tag
	feed         repo.Feed
	subscription repo.Subscription
	article      repo.Article
	extract      repo.Extract
	scores       repo.Scores
	thumbnail    repo.Thumbnail
}

func NewService(source string, log log.Log) (Service, error) {
	if err := os.MkdirAll(filepath.Dir(source), 0700); err != nil {
		return Service{}, errors.Wrap(err, "creating source directory")
	}

	db, err := storm.Open(source, storm.Codec(msgpack.Codec))
	if err != nil {
		return Service{}, errors.Wrap(err, "opening content database")
	}

	userRepo, err := newUserRepo(db, log)
	if err != nil {
		return Service{}, err
	}

	feedRepo, err := newFeedRepo(db, log)
	if err != nil {
		return Service{}, err
	}

	return Service{
		user: userRepo,
		feed: feedRepo,
	}, nil
}

func (s Service) UserRepo() repo.User {
	return s.user
}

func (s Service) TagRepo() repo.Tag {
	return s.tag
}

func (s Service) FeedRepo() repo.Feed {
	return s.feed
}

func (s Service) SubscriptionRepo() repo.Subscription {
	return s.subscription
}

func (s Service) ArticleRepo() repo.Article {
	return s.article
}

func (s Service) ExtractRepo() repo.Extract {
	return s.extract
}

func (s Service) ScoresRepo() repo.Scores {
	return s.scores
}

func (s Service) ThumbnailRepo() repo.Thumbnail {
	return s.thumbnail
}
