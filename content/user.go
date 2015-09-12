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

	ArticleById(id data.ArticleId) UserArticle

	ArticlesById(ids []data.ArticleId) []UserArticle

	AllUnreadArticleIds() []data.ArticleId
	AllFavoriteArticleIds() []data.ArticleId

	Tags() []Tag
}

type UserRelated interface {
	User(u ...User) User
}
