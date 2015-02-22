package sql

import (
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/info"
	"github.com/urandom/readeef/db"
	"github.com/urandom/webfw"
)

type Feed struct {
	base.Feed
	NamedSQL
	logger webfw.Logger

	db *db.DB
}

type UserFeed struct {
	base.UserFeed
	Feed
}

type TaggedFeed struct {
	UserFeed
}

func NewFeed(db *db.DB, logger webfw.Logger) Feed {
	f := Feed{NamedSQL: NewNamedSQL(), db: db, logger: logger}

	f.init()

	return f
}

func NewUserFeed(db *db.DB, logger webfw.Logger, user content.User) UserFeed {
	uf := UserFeed{Feed: NewFeed(db, logger), UserFeed: base.NewUserFeed(user)}

	uf.init()

	return uf
}

func (f Feed) init() {
}

func (uf *UserFeed) Set(info info.Feed) {
	uf.Feed.Set(info)
}

func (uf UserFeed) init() {
}
