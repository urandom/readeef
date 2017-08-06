package repo

import "github.com/urandom/readeef/content"

// Extract allows fetching and manipulating content.Extract objects
type Extract interface {
	Get(content.Article) (content.Extract, error)
	Update(content.Extract) error
}
