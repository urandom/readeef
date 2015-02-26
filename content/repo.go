package content

import "github.com/urandom/readeef/content/info"

type Repo interface {
	Error
	Generator

	UserByLogin(login info.Login) User
	UserByMD5Api(md5 []byte) User
	AllUsers() []User

	FeedById(id info.FeedId) Feed
	FeedByLink(link string) Feed

	AllFeeds() []Feed
	AllUnsubscribedFeeds() []Feed

	AllSubscriptions() []Subscription

	FailSubscriptions()
}

type Generator interface {
	Article() Article
	UserArticle(u User) UserArticle
	ScoredArticle(u User) ScoredArticle

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
