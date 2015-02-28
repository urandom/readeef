package content

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/urandom/readeef/content/info"
)

type User interface {
	Error
	ArticleSorting
	RepoRelated
	ArticleSearch

	fmt.Stringer
	json.Marshaler

	Info(info ...info.User) info.User

	Validate() error

	Password(password string, secret []byte)
	Authenticate(password string, secret []byte) bool

	Update()
	Delete()

	FeedById(id info.FeedId) UserFeed
	AddFeed(feed Feed) UserFeed

	AllFeeds() []UserFeed

	AllTaggedFeeds() []TaggedFeed

	ArticleById(id info.ArticleId) UserArticle

	ArticlesById(ids []info.ArticleId) []UserArticle

	AllUnreadArticleIds() []info.ArticleId
	AllFavoriteIds() []info.ArticleId

	ArticleCount() int64

	Articles(paging ...int) []UserArticle
	UnreadArticles(paging ...int) []UserArticle
	ArticlesOrderedById(pivot info.ArticleId, paging ...int) []UserArticle
	FavoriteArticles(paging ...int) []UserArticle

	ReadBefore(date time.Time, read bool)
	ReadAfter(date time.Time, read bool)

	ScoredArticles(from, to time.Time, paging ...int) []ScoredArticle

	Tags() []Tag
}

type UserRelated interface {
	User(u ...User) User
}
