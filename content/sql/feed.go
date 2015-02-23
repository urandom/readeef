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

func (f *Feed) Update(info info.Feed) content.Feed {
	if f.Err() != nil {
		return f
	}

	return f
}

func (f *Feed) Delete() content.Feed {
	if f.Err() != nil {
		return f
	}

	return f
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

func (f *Feed) AddArticles([]content.Article) content.Feed {
	if f.Err() != nil {
		return f
	}

	return f
}

func (f *Feed) Subscription() (s content.Subscription) {
	if f.Err() != nil {
		return
	}

	return
}

func (f *Feed) init() {
}

func (uf UserFeed) Err() error {
	return uf.Feed.Err()
}

func (uf *UserFeed) SetErr(err error) content.Error {
	return uf.Feed.SetErr(err)
}

func (uf *UserFeed) DefaultSorting() content.ArticleSorting {
	return uf.Feed.DefaultSorting()
}

func (uf *UserFeed) SortingById() content.ArticleSorting {
	return uf.Feed.SortingById()
}

func (uf *UserFeed) SortingByDate() content.ArticleSorting {
	return uf.Feed.SortingByDate()
}

func (uf *UserFeed) Reverse() content.ArticleSorting {
	return uf.Feed.Reverse()
}

func (uf *UserFeed) Set(info info.Feed) content.Feed {
	return uf.Feed.Set(info)
}

func (uf UserFeed) Info() info.Feed {
	return uf.Feed.Info()
}

func (uf UserFeed) String() string {
	return uf.Feed.String()
}

func (uf *UserFeed) Users() (u []content.User) {
	if uf.Err() != nil {
		return
	}

	return
}

func (uf *UserFeed) Detach() content.UserFeed {
	if uf.Err() != nil {
		return uf
	}

	return uf
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

func (uf *UserFeed) ReadBefore(date time.Time, read bool) content.UserFeed {
	if uf.Err() != nil {
		return uf
	}

	return uf
}

func (uf *UserFeed) ScoredArticles(from, to time.Time, paging ...int) (sa []content.ScoredArticle) {
	if uf.Err() != nil {
		return
	}

	return
}

func (uf *UserFeed) init() {
}

func (tf *TaggedFeed) AddTags(tags ...content.Tag) content.TaggedFeed {
	if tf.Err() != nil {
		return tf
	}

	return tf
}

func (tf *TaggedFeed) DeleteAllTags() content.TaggedFeed {
	if tf.Err() != nil {
		return tf
	}

	return tf
}

func (tf *TaggedFeed) Highlight(highlight string) content.ArticleSearch {
	return tf.UserFeed.Highlight(highlight)
}

func (tf *TaggedFeed) Query(query string) []content.UserArticle {
	return tf.UserFeed.Query(query)
}

func (tf TaggedFeed) User() content.User {
	return tf.UserFeed.User()
}

func (tf TaggedFeed) Validate() error {
	return tf.UserFeed.Validate()
}

func (tf *TaggedFeed) init() {
}
