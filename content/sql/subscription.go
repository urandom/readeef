package sql

import (
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/info"
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

func (s *Subscription) Update(info info.Subscription) {
	if s.Err() != nil {
		return
	}
}

func (s *Subscription) Delete() {
	if s.Err() != nil {
		return
	}
}

func (s *Subscription) Fail(fail bool) {
	if s.Err() != nil {
		return
	}
}

func init() {
}
