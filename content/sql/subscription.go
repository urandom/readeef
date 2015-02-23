package sql

import (
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/info"
	"github.com/urandom/readeef/db"
	"github.com/urandom/webfw"
)

type Subscription struct {
	base.Subscription
	NamedSQL
	logger webfw.Logger

	db *db.DB
}

func NewSubscription(db *db.DB, logger webfw.Logger) *Subscription {
	s := &Subscription{NamedSQL: NewNamedSQL(), db: db, logger: logger}

	s.init()

	return s
}

func (s *Subscription) Update(info info.Subscription) content.Subscription {
	if s.Err() != nil {
		return s
	}

	return s
}

func (s *Subscription) Delete() content.Subscription {
	if s.Err() != nil {
		return s
	}

	return s
}

func (s *Subscription) Fail(fail bool) content.Subscription {
	if s.Err() != nil {
		return s
	}

	return s
}

func (s *Subscription) init() {
}
