package content

import (
	"fmt"
	"time"

	"github.com/urandom/readeef/content/info"
)

type Feed interface {
	Error
	ArticleSorting

	fmt.Stringer

	Set(info info.Feed) Feed
	Info() info.Feed

	Validate() error

	// Returns any new articles since the previous Update
	NewArticles() []Article

	Update(info info.Feed) Feed
	Delete() Feed

	AllArticles() []Article
	LatestArticles() []Article

	AddArticles([]Article) Feed

	Subscription() Subscription
}

type UserFeed interface {
	Feed
	ArticleSearch

	User() User

	Users() []User

	// Detaches from the current user
	Detach() UserFeed

	Articles(desc bool, paging ...int) []UserArticle
	UnreadArticles(desc bool, paging ...int) []UserArticle

	ReadBefore(date time.Time, read bool) UserFeed

	ScoredArticles(from, to time.Time, paging ...int) []ScoredArticle
}

type TaggedFeed interface {
	UserFeed

	Tags() []Tag
	SetTags(tags []Tag) TaggedFeed

	AddTags(tags ...Tag) TaggedFeed
	DeleteAllTags() TaggedFeed
}
