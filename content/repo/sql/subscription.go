package sql

import (
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/sql/db"
	"github.com/urandom/readeef/log"
)

type subscriptionRepo struct {
	db *db.DB

	log log.Log
}

func (r subscriptionRepo) All() ([]content.Subscription, error) {
	r.log.Infoln("Getting all subscriptions")

	var subscriptions []content.Subscription
	err := r.db.Select(&subscriptions, r.db.SQL().Repo.GetHubbubSubscriptions)
	if err != nil {
		return []content.Subscription{}, errors.Wrap(err, "getting hubbub subscriptions")
	}

	return subscriptions, nil
}

func (r subscriptionRepo) Update(content.Subscription) error {
	panic("not implemented")
}
