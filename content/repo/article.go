package repo

import "github.com/urandom/readeef/content"

// Article allows fetching and manipulating content.Article objects
type Article interface {
	ForUser(content.User, ...content.QueryOpt) ([]content.Article, error)
	All(...content.QueryOpt) ([]content.Article, error)

	Read(content.Article, bool, content.User, ...content.Article) error
	Favorite(content.Article, bool, content.User, ...content.Article) error

	RemoveStaleUnreadRecords() error
}
