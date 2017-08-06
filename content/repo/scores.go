package repo

import "github.com/urandom/readeef/content"

// Scores allows fetching and manipulating content.Scores objects
type Scores interface {
	Get(content.Article) (content.Scores, error)
	Update(content.Scores) error
}
