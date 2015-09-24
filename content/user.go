package content

import (
	"encoding/json"
	"fmt"

	"github.com/urandom/readeef/content/data"
)

type User interface {
	Error
	RepoRelated
	ArticleSearch
	ArticleRepo

	fmt.Stringer
	json.Marshaler

	Data(data ...data.User) data.User

	Validate() error

	Password(password string, secret []byte)
	Authenticate(password string, secret []byte) bool

	Update()
	Delete()

	FeedById(id data.FeedId) UserFeed
	AddFeed(feed Feed) UserFeed

	AllFeeds() []UserFeed

	AllTaggedFeeds() []TaggedFeed

	ArticleById(id data.ArticleId, opts ...data.ArticleQueryOptions) UserArticle
	ArticlesById(ids []data.ArticleId, opts ...data.ArticleQueryOptions) []UserArticle

	Tags() []Tag
	TagById(id data.TagId) Tag
	TagByValue(v data.TagValue) Tag
}

type UserRelated interface {
	User(u ...User) User
}
