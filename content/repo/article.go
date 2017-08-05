package repo

import "github.com/urandom/readeef/content"

// Article allows fetching and manipulating content.Article objects
type Article interface {
	ForUser(content.User, ...content.QueryOpt) ([]content.Article, error)

	All(...content.QueryOpt) ([]content.Article, error)

	Count(content.User, ...content.QueryOpt) (int64, error)
	IDs(content.User, ...content.QueryOpt) ([]content.ArticleID, error)

	Read(bool, content.User, ...content.QueryOpt) error
	Favor(bool, content.User, ...content.QueryOpt) error

	RemoveStaleUnreadRecords() error
}
