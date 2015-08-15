package content

import "github.com/urandom/readeef/content/data"

type Repo interface {
	Error
	Generator

	UserByLogin(login data.Login) User
	UserByMD5Api(md5 []byte) User
	AllUsers() []User

	FeedById(id data.FeedId) Feed
	FeedByLink(link string) Feed

	AllFeeds() []Feed
	AllUnsubscribedFeeds() []Feed

	AllSubscriptions() []Subscription

	FailSubscriptions()
}

type Generator interface {
	Article() Article
	ScoredArticle() ScoredArticle
	UserArticle(u User) UserArticle

	ArticleThumbnail() ArticleThumbnail
	ArticleScores() ArticleScores

	Feed() Feed
	UserFeed(u User) UserFeed
	TaggedFeed(u User) TaggedFeed

	Subscription() Subscription

	Tag(u User) Tag

	User() User

	Domain(u string) Domain
}

type RepoRelated interface {
	Repo(r ...Repo) Repo
}
