package content

import (
	"fmt"
	"time"

	"github.com/urandom/readeef/content/info"
	"github.com/urandom/readeef/parser"
)

type Feed interface {
	Error
	ArticleSorting

	fmt.Stringer

	Info(in ...info.Feed) info.Feed

	Validate() error

	Users() []User

	// Updates the in-memory feed information using the RSS data
	Refresh(pf parser.Feed)
	// Returns the []content.Article created from the RSS data
	ParsedArticles() []Article

	// Returns any new articles since the previous Update
	NewArticles() []Article

	Update()
	Delete()

	AllArticles() []Article
	LatestArticles() []Article

	AddArticles([]Article)

	Subscription() Subscription
}

type UserFeed interface {
	Feed
	ArticleSearch
	UserRelated
	RepoRelated

	// Detaches from the current user
	Detach()

	Articles(paging ...int) []UserArticle
	UnreadArticles(paging ...int) []UserArticle

	ReadBefore(date time.Time, read bool)

	ScoredArticles(from, to time.Time, paging ...int) []ScoredArticle
}

type TaggedFeed interface {
	UserFeed

	Tags(tags ...[]Tag) []Tag

	AddTags(tags ...Tag)
	DeleteAllTags()
}
