package repo

import "github.com/urandom/readeef/content"

// Feed allows fetching and manipulating content.Feed objects
type Feed interface {
	Get(content.FeedID) (content.Feed, error)
	FindByLink(link string) (content.Feed, error)

	IDs() ([]content.FeedID, error)
	Unsubscribed() ([]content.Feed, error)

	Update(content.Feed) ([]content.Article, error)
	Delete(content.Feed) error

	Users(content.Feed) ([]content.User, error)
	DetachFrom(content.Feed, content.User) error

	Tags(content.Feed) ([]content.Tag, error)
	UpdateTags(content.Feed, []content.Tag) error
}
