package content

import "github.com/urandom/readeef/content/data"

type Repo interface {
	Error
	Generator

	ArticleProcessors(processors ...[]ArticleProcessor) []ArticleProcessor

	UserByLogin(login data.Login) User
	UserByMD5Api(md5 []byte) User
	AllUsers() []User

	FeedById(id data.FeedId) Feed
	FeedByLink(link string) Feed

	AllFeeds() []Feed
	AllUnsubscribedFeeds() []Feed

	AllSubscriptions() []Subscription
	FailSubscriptions()

	DeleteStaleUnreadRecords()
}

type Generator interface {
	Article() Article
	UserArticle(u User) UserArticle

	ArticleThumbnail() ArticleThumbnail
	ArticleExtract() ArticleExtract
	ArticleScores() ArticleScores

	Feed() Feed
	UserFeed(u User) UserFeed
	TaggedFeed(u User) TaggedFeed

	Subscription() Subscription

	Tag(u User) Tag

	User() User
}

type RepoRelated interface {
	Repo(r ...Repo) Repo
}
