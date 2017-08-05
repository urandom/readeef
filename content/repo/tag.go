package repo

import "github.com/urandom/readeef/content"

// Tag allows fetching and manipulating content.Tag objects
type Tag interface {
	Get(content.TagID, content.User) (content.Tag, error)

	ForUser(content.User) ([]content.Tag, error)
	ForFeed(content.Feed, content.User) ([]content.Tag, error)

	FeedIDs(content.Tag, content.User) ([]content.FeedID, error)
}
