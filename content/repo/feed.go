package repo

import "github.com/urandom/readeef/content"

// Feed allows fetching and manipulating content.Feed objects
type Feed interface {
	IDs() ([]content.FeedID, error)

	Update(content.Feed) ([]content.Article, error)
	Delete(content.Feed) error

	Users(content.Feed) ([]content.User, error)
	DetachFrom(content.Feed, content.User) error

	Tags(content.Feed) ([]content.Tag, error)
	UpdateTags(content.Feed, []content.Tag) error
}
