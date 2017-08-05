package repo

import "github.com/urandom/readeef/content"

// User allows manipulating content.User objects
type User interface {
	Get(content.Login) (content.User, error)
	All() ([]content.User, error)
	Update(content.User) error
	Delete(content.User) error

	FindByMD5([]byte) (content.User, error)
}
