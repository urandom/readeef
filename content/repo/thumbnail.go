package repo

import "github.com/urandom/readeef/content"

// Thumbnail allows fetching and manipulating content.Thumbnail objects
type Thumbnail interface {
	Get(content.ArticleID) (content.Thumbnail, error)
	Update(content.Thumbnail) error
}
