package content

import (
	"fmt"

	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/parser"
)

type Feed interface {
	Error

	fmt.Stringer

	Data(data ...data.Feed) data.Feed

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

	SetNewArticlesUnread()

	AllArticles() []Article
	LatestArticles() []Article

	AddArticles([]Article)

	Subscription() Subscription
}

type UserFeed interface {
	Feed
	ArticleSearch
	ArticleRepo
	RepoRelated

	// Detaches from the current user
	Detach()
}

type TaggedFeed interface {
	UserFeed

	Tags(tags ...[]Tag) []Tag

	UpdateTags()
}
