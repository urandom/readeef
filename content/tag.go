package content

import (
	"encoding/json"
	"fmt"

	"github.com/urandom/readeef/content/data"
)

type Tag interface {
	Error
	ArticleSearch
	ArticleRepo
	RepoRelated

	fmt.Stringer
	json.Marshaler

	Data(data ...data.Tag) data.Tag

	Validate() error

	AllFeeds() []TaggedFeed
}
