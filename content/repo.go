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
