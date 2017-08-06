package sql

import (
	"database/sql"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/sql/db"
	"github.com/urandom/readeef/log"
)

type subscriptionRepo struct {
	db *db.DB

	log log.Log
}

func (r subscriptionRepo) Get(feed content.FeedID) (content.Subscription, error) {
	if err := feed.Validate(); err != nil {
		return content.Subscription{}, errors.WithMessage("validating feed")
	}
	r.log.Infoln("Getting feed subscription")

	var subscription content.Subscription
	err := r.db.Get(subscription, r.db.SQL().Feed.GetHubbubSubscription, feed.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}

		return content.Subscription{}, errors.Wrapf(err, "getting subscription for feed %s", feed)
	}

	return subscription, nil
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

func (r subscriptionRepo) Update(s content.Subscription) error {
	if err := s.Validate(); err != nil {
		return errors.WithMessage(err, "validating subscription")
	}

	r.log.Infof("Updating subscription %s", s)

	tx, err := r.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "creating transaction")
	}
	defer tx.Rollback()

	s := r.db.SQL()
	stmt, err := tx.Preparex(s.Subscription.Update)
	if err != nil {
		return errors.Wrap(err, "preparing subscription update stmt")
	}
	defer stmt.Close()

	res, err := stmt.Exec(s.FeedID, s.LeaseDuration, s.VerificationTime, s.SubscriptionFailure, s.Link)
	if err != nil {
		return errors.Wrap(err, "executimg subscription update stmt")
	}

	if num, err := res.RowsAffected(); err == nil && num > 0 {
		if err := tx.Commit(); err != nil {
			return errors.Wrap(err, "committing transaction")
		}

		return nil
	}

	stmt, err = tx.Preparex(s.Subscription.Create)
	if err != nil {
		return errors.Wrap(err, "preparing subscription create stmt")
	}
	defer stmt.Close()

	_, err = stmt.Exec(s.Link, s.FeedID, s.LeaseDuration, s.VerificationTime, s.SubscriptionFailure)
	if err != nil {
		return errors.Wrap(err, "executimg subscription create stmt")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "committing transaction")
	}

	return nil

}
