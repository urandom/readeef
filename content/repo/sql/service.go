package sql

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/content/repo/sql/db"
	"github.com/urandom/readeef/log"
)

type Service struct {
	db *db.DB

	log log.Log
}

func NewService(driver, source string, log log.Log) (Service, error) {
	switch driver {
	case "sqlite3", "postgres":
		db := db.New(log)
		if err := db.Open(driver, source); err != nil {
			return Service{}, errors.Wrap(err, "connecting to database")
		}

		return Service{db, log}, nil
	default:
		panic(fmt.Sprintf("Cannot provide a repo for driver '%s'\n", driver))
	}
}

func (s Service) UserRepo() repo.User {
	return userRepo{s.db, s.log}
}

func (s Service) TagRepo() repo.Tag {
	return tagRepo{s.db, s.log}
}

func (s Service) FeedRepo() repo.Feed {
	return feedRepo{s.db, s.log}
}

func (s Service) SubscriptionRepo() repo.Subscription {
	return subscriptionRepo{s.db, s.log}
}

func (s Service) ArticleRepo() repo.Article {
	return articleRepo{s.db, s.log}
}

func (s Service) ExtractRepo() repo.Extract {
	return extractRepo{s.db, s.log}
}

func (s Service) ThumbnailRepo() repo.Thumbnail {
	return thumbnailRepo{s.db, s.log}
}

func (s Service) ScoresRepo() repo.Scores {
	return scoresRepo{s.db, s.log}
}
