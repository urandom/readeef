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

	Set(info info.Feed)
	Info() info.Feed

	Validate()

	// Returns any new articles since the previous Update
	NewArticles() []Article

	Update(info info.Feed)
	Delete()

	AllArticles() []Article
	LatestArticles() []Article

	AddArticles([]Article)

	Subscription() Subscription
}

type UserFeed interface {
	Feed
	ArticleSearch

	User() User

	Users() []User

	// Detaches from the current user
	Detach()

	Articles(desc bool, paging ...int) []UserArticle
	UnreadArticles(desc bool, paging ...int) []UserArticle

	ReadBefore(date time.Time, read bool)

	ScoredArticles(from, to time.Time, paging ...int) []ScoredArticle
}

type TaggedFeed interface {
	UserFeed

	Tags() []Tag

	AddTags(tags ...Tag)
	DeleteAllTags()
}
