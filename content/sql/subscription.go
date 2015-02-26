package sql

import (
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/db"
	"github.com/urandom/webfw"
)

type Subscription struct {
	base.Subscription
	logger webfw.Logger

	db *db.DB
}

func NewSubscription(db *db.DB, logger webfw.Logger) *Subscription {
	return &Subscription{db: db, logger: logger}
}

func (s *Subscription) Update() {
	if s.Err() != nil {
		return
	}

	i := s.Info()
	s.logger.Infof("Updating subscription to %s\n", i.Link)

	tx, err := s.db.Begin()
	if err != nil {
		s.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.SQL("update_hubbub_subscription"))
	if err != nil {
		s.Err(err)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(i.FeedId, i.LeaseDuration, i.VerificationTime, i.SubscriptionFailure, i.Link)
	if err != nil {
		s.Err(err)
		return
	}

	if num, err := res.RowsAffected(); err == nil && num > 0 {
		tx.Commit()
		return
	}

	stmt, err = tx.Preparex(db.SQL("create_hubbub_subscription"))

	if err != nil {
		s.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(i.Link, i.FeedId, i.LeaseDuration, i.VerificationTime, i.SubscriptionFailure)
	if err != nil {
		s.Err(err)
		return
	}

	tx.Commit()
}

func (s *Subscription) Delete() {
	if s.Err() != nil {
		return
	}

	i := s.Info()
	s.logger.Infof("Deleting subscription to %s\n", i.Link)

	tx, err := s.db.Begin()
	if err != nil {
		s.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.SQL("delete_hubbub_subscription"))

	if err != nil {
		s.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(i.Link)
	if err != nil {
		s.Err(err)
		return
	}

	tx.Commit()
}

func init() {
	db.SetSQL("create_hubbub_subscription", createHubbubSubscription)
	db.SetSQL("update_hubbub_subscription", updateHubbubSubscription)
	db.SetSQL("delete_hubbub_subscription", deleteHubbubSubscription)
}

const (
	createHubbubSubscription = `
INSERT INTO hubbub_subscriptions(link, feed_id, lease_duration, verification_time, subscription_failure)
	SELECT $1, $2, $3, $4, $5 EXCEPT
	SELECT link, feed_id, lease_duration, verification_time, subscription_failure
		FROM hubbub_subscriptions WHERE link = $1
`
	updateHubbubSubscription = `
UPDATE hubbub_subscriptions SET feed_id = $1, lease_duration = $2,
	verification_time = $3, subscription_failure = $4 WHERE link = $5
`
	deleteHubbubSubscription = `DELETE from hubbub_subscriptions where link = $1`
)
