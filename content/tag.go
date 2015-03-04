package content

import (
	"encoding/json"
	"fmt"

	"github.com/urandom/readeef/content/data"
)

type Tag interface {
	Error
	ArticleSorting
	ArticleSearch
	ArticleRepo
	RepoRelated

	fmt.Stringer
	json.Marshaler

	Value(val ...data.TagValue) data.TagValue

	Validate() error

	AllFeeds() []TaggedFeed
}
