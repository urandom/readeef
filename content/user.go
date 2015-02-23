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

	Set(info info.User)
	Info() info.User

	Validate() error

	Password(password string, secret []byte)
	Authenticate(password string, secret []byte) bool

	Update()
	Delete()

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

	ReadBefore(date time.Time, read bool)
	ReadAfter(date time.Time, read bool)

	ScoredArticles(from, to time.Time, desc bool, paging ...int) []ScoredArticle

	Tags() []Tag
}
