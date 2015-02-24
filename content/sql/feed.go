package sql

import (
	"time"

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

	db          *db.DB
	newArticles []content.Article
}

type UserFeed struct {
	base.UserFeed
	*Feed
}

type TaggedFeed struct {
	base.TaggedFeed
	*UserFeed
}

func NewFeed(db *db.DB, logger webfw.Logger) *Feed {
	f := &Feed{NamedSQL: NewNamedSQL(), db: db, logger: logger}

	f.init()

	return f
}

func NewUserFeed(db *db.DB, logger webfw.Logger, user content.User) *UserFeed {
	uf := &UserFeed{Feed: NewFeed(db, logger), UserFeed: base.NewUserFeed(user)}

	uf.init()

	return uf
}

func NewTaggedFeed(db *db.DB, logger webfw.Logger, user content.User) *TaggedFeed {
	tf := &TaggedFeed{UserFeed: NewUserFeed(db, logger, user)}

	tf.init()

	return tf
}

func (f Feed) NewArticles() (a []content.Article) {
	if f.Err() != nil {
		return
	}

	return f.newArticles
}

func (f *Feed) Update(info info.Feed) {
	if f.Err() != nil {
		return
	}
}

func (f *Feed) Delete() {
	if f.Err() != nil {
		return
	}
}

func (f *Feed) AllArticles() (a []content.Article) {
	if f.Err() != nil {
		return
	}

	return
}

func (f *Feed) LatestArticles() (a []content.Article) {
	if f.Err() != nil {
		return
	}

	return
}

func (f *Feed) AddArticles([]content.Article) {
	if f.Err() != nil {
		return
	}
}

func (f *Feed) Subscription() (s content.Subscription) {
	if f.Err() != nil {
		return
	}

	return
}

func (f *Feed) init() {
}

func (uf UserFeed) Validate() error {
	err := uf.Feed.Validate()
	if err == nil {
		err = uf.UserFeed.Validate()
	}

	return err
}

func (uf *UserFeed) Users() (u []content.User) {
	if uf.Err() != nil {
		return
	}

	return
}

func (uf *UserFeed) Detach() {
	if uf.Err() != nil {
		return
	}
}

func (uf *UserFeed) Articles(desc bool, paging ...int) (ua []content.UserArticle) {
	if uf.Err() != nil {
		return
	}

	return
}

func (uf *UserFeed) UnreadArticles(desc bool, paging ...int) (ua []content.UserArticle) {
	if uf.Err() != nil {
		return
	}

	return
}

func (uf *UserFeed) ReadBefore(date time.Time, read bool) {
	if uf.Err() != nil {
		return
	}
}

func (uf *UserFeed) ScoredArticles(from, to time.Time, paging ...int) (sa []content.ScoredArticle) {
	if uf.Err() != nil {
		return
	}

	return
}

func (uf *UserFeed) init() {
}

func (tf *TaggedFeed) Tags() (t []content.Tag) {
	if tf.Err() != nil {
		return
	}

	return
}

func (tf *TaggedFeed) AddTags(tags ...content.Tag) {
	if tf.Err() != nil {
		return
	}
}

func (tf *TaggedFeed) DeleteAllTags() {
	if tf.Err() != nil {
		return
	}
}

func (tf *TaggedFeed) init() {
}
