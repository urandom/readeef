package repo

import "github.com/urandom/readeef/content"

// Subscription allows fetching and manipulating content.Subscription objects
type Subscription interface {
	Get(content.Feed) (content.Subscription, error)
	All() ([]content.Subscription, error)

	Update(content.Subscription) error
}
