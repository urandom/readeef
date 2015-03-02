package content

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/urandom/readeef/content/data"
)

type Tag interface {
	Error
	ArticleSorting
	ArticleSearch
	RepoRelated

	fmt.Stringer
	json.Marshaler

	Value(val ...data.TagValue) data.TagValue

	Validate() error

	AllFeeds() []TaggedFeed

	Articles(paging ...int) []UserArticle
	UnreadArticles(paging ...int) []UserArticle

	ReadBefore(date time.Time, read bool)

	ScoredArticles(from, to time.Time, paging ...int) []ScoredArticle
}
