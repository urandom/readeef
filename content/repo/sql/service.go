package sql

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/content/sql/db"
)

type Service struct {
	db *db.DB

	log readeef.Logger
}

func NewService(driver, source string, log readeef.Logger) (Service, error) {
	switch driver {
	case "sqlite3", "postgres":
		db := db.New(log)
		if err = db.Open(driver, source); err != nil {
			return Service{}, errors.Wrap(err, "connecting to database")
		}

		return Service{db, log}, nil
	default:
		panic(fmt.Sprintf("Cannot provide a repo for driver '%s'\n", driver))
	}
}

func (s Service) UserRepo() repo.User {
	return userRepo{db, log}
}

func (s Service) FeedRepo() repo.Feed {
	return feedRepo{db, log}
}

func (s Service) SubscriptionRepo() repo.Subscription {
	return subscriptionRepo{db, log}
}

func (s Service) ArticleRepo() repo.Article {
	return articleRepo{db, log}
}
