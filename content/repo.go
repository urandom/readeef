package content

import "github.com/urandom/readeef/content/info"

type Repo interface {
	Error

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
