package content

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/urandom/readeef/content/data"
)

type User interface {
	Error
	ArticleSorting
	RepoRelated
	ArticleSearch

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
	AllFavoriteIds() []data.ArticleId

	ArticleCount() int64

	Articles(paging ...int) []UserArticle
	UnreadArticles(paging ...int) []UserArticle
	ArticlesOrderedById(pivot data.ArticleId, paging ...int) []UserArticle
	FavoriteArticles(paging ...int) []UserArticle

	ReadBefore(date time.Time, read bool)
	ReadAfter(date time.Time, read bool)

	ScoredArticles(from, to time.Time, paging ...int) []ScoredArticle

	Tags() []Tag
}

type UserRelated interface {
	User(u ...User) User
}
