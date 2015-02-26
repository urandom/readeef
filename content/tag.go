package content

import (
	"fmt"
	"time"

	"github.com/urandom/readeef/content/info"
)

type Tag interface {
	Error
	ArticleSorting
	ArticleSearch
	UserRelated
	RepoRelated

	fmt.Stringer

	Set(value info.TagValue)
	Value() info.TagValue

	AllFeeds() []TaggedFeed

	Articles(paging ...int) []UserArticle
	UnreadArticles(paging ...int) []UserArticle

	ReadBefore(date time.Time, read bool)

	ScoredArticles(from, to time.Time, paging ...int) []ScoredArticle
}
