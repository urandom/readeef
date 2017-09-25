package repo

import "github.com/urandom/readeef/content"

// Feed allows fetching and manipulating content.Feed objects
type Feed interface {
	Get(content.FeedID, content.User) (content.Feed, error)
	FindByLink(link string) (content.Feed, error)

	ForUser(content.User) ([]content.Feed, error)
	ForTag(content.Tag, content.User) ([]content.Feed, error)
	All() ([]content.Feed, error)

	IDs() ([]content.FeedID, error)
	Unsubscribed() ([]content.Feed, error)

	Update(*content.Feed) ([]content.Article, error)
	Delete(content.Feed) error

	Users(content.Feed) ([]content.User, error)
	AttachTo(content.Feed, content.User) error
	DetachFrom(content.Feed, content.User) error

	SetUserTags(content.Feed, content.User, []*content.Tag) error
}
