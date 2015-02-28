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
	RepoRelated

	fmt.Stringer

	Value(val ...info.TagValue) info.TagValue

	Validate() error

	AllFeeds() []TaggedFeed

	Articles(paging ...int) []UserArticle
	UnreadArticles(paging ...int) []UserArticle

	ReadBefore(date time.Time, read bool)

	ScoredArticles(from, to time.Time, paging ...int) []ScoredArticle
}
