package content

import (
	"fmt"
	"time"

	"github.com/urandom/readeef/content/info"
)

type User interface {
	Error
	ArticleSorting

	fmt.Stringer

	Set(info info.User) User
	Info() info.User

	Validate() error

	Password(password string, secret []byte) User
	Authenticate(password string, secret []byte) bool

	Update() User
	Delete() User

	Feed(id info.FeedId) UserFeed
	AddFeed(feed Feed) UserFeed

	AllFeeds() []UserFeed

	AllTaggedFeeds() []TaggedFeed

	Article(id info.ArticleId) UserArticle

	ArticlesById(ids ...info.ArticleId) []UserArticle

	AllUnreadArticleIds() []info.ArticleId
	AllFavoriteIds() []info.ArticleId

	ArticleCount() int64

	Articles(desc bool, paging ...int) []UserArticle
	UnreadArticles(desc bool, paging ...int) []UserArticle
	ArticlesOrderedById(pivot info.ArticleId, desc bool, paging ...int) []UserArticle
	FavoriteArticles(desc bool, paging ...int) []UserArticle

	ReadBefore(date time.Time, read bool) User
	ReadAfter(date time.Time, read bool) User

	ScoredArticles(from, to time.Time, desc bool, paging ...int) []ScoredArticle

	Tags() []Tag
}
