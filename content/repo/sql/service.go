package sql

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/content/repo/sql/db"
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

func NewService(driver, source string, log log.Log) (Service, error) {
	switch driver {
	case "sqlite3", "postgres":
		db := db.New(log)
		if err := db.Open(driver, source); err != nil {
			return Service{}, errors.Wrap(err, "connecting to database")
		}

		return Service{
			user:         userRepo{db, log},
			tag:          tagRepo{db, log},
			feed:         feedRepo{db, log},
			subscription: subscriptionRepo{db, log},
			article:      articleRepo{db, log},
			extract:      extractRepo{db, log},
			scores:       scoresRepo{db, log},
			thumbnail:    thumbnailRepo{db, log},
		}, nil
	default:
		panic(fmt.Sprintf("Cannot provide a repo for driver '%s'\n", driver))
	}
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
