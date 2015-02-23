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

	fmt.Stringer

	Set(value info.TagValue) Tag
	Value() info.TagValue

	AllFeeds() []TaggedFeed

	Articles(desc bool, paging ...int) []UserArticle
	UnreadArticles(desc bool, paging ...int) []UserArticle

	ReadBefore(date time.Time, read bool) Tag

	ScoredArticles(from, to time.Time, desc bool, paging ...int) []ScoredArticle
}
