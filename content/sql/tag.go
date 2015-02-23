package sql

import (
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/db"
	"github.com/urandom/webfw"
)

type Tag struct {
	base.Tag
	NamedSQL
	logger webfw.Logger

	db *db.DB
}

func NewTag(db *db.DB, logger webfw.Logger) *Tag {
	t := &Tag{NamedSQL: NewNamedSQL(), db: db, logger: logger}

	t.init()

	return t
}

func (t *Tag) AllFeeds() (tf []content.TaggedFeed) {
	if t.Err() != nil {
		return
	}

	return
}

func (t *Tag) Articles(desc bool, paging ...int) (ua []content.UserArticle) {
	if t.Err() != nil {
		return
	}

	return
}

func (t *Tag) UnreadArticles(desc bool, paging ...int) (ua []content.UserArticle) {
	if t.Err() != nil {
		return
	}

	return
}

func (t *Tag) ReadBefore(date time.Time, read bool) {
	if t.Err() != nil {
		return
	}
}

func (t *Tag) ScoredArticles(from, to time.Time, desc bool, paging ...int) (sa []content.ScoredArticle) {
	if t.Err() != nil {
		return
	}

	return
}

func (t *Tag) init() {
}
