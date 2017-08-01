package repo

import "github.com/urandom/readeef/content"

// Article allows fetching and manipulating content.Article objects
type Article interface {
	ForUser(content.User, ...content.QueryOpt) ([]content.Article, error)
	All(...content.QueryOpt) ([]content.Article, error)
}
