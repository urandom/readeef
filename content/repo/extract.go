package repo

import "github.com/urandom/readeef/content"

// Extract allows fetching and manipulating content.Extract objects
type Extract interface {
	Get(content.ArticleID) (content.Extract, error)
	Update(content.Extract) error
}
