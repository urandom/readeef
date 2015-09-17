package sql

import (
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/sql/db"
	"github.com/urandom/webfw"
)

type Subscription struct {
	base.Subscription
	logger webfw.Logger

	db *db.DB
}

func (s *Subscription) Update() {
	if s.HasErr() {
		return
	}

	if err := s.Validate(); err != nil {
		s.Err(err)
		return
	}

	i := s.Data()
	s.logger.Infof("Updating subscription to %s\n", i.Link)

	tx, err := s.db.Beginx()
	if err != nil {
		s.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(s.db.SQL().Subscription.Update)
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

	stmt, err = tx.Preparex(s.db.SQL().Subscription.Create)

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
	if s.HasErr() {
		return
	}

	if err := s.Validate(); err != nil {
		s.Err(err)
		return
	}

	i := s.Data()
	s.logger.Infof("Deleting subscription to %s\n", i.Link)

	tx, err := s.db.Beginx()
	if err != nil {
		s.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(s.db.SQL().Subscription.Delete)

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
