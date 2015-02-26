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

	stmt, err := tx.Preparex(s.db.SQL("update_hubbub_subscription"))
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

	stmt, err = tx.Preparex(s.db.SQL("create_hubbub_subscription"))

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

	stmt, err := tx.Preparex(s.db.SQL("delete_hubbub_subscription"))

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
