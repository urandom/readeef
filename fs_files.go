// +build !nofs

// DO NOT EDIT ** This file was generated with github.com/urandom/embed ** DO NOT EDIT //

package readeef

import (
	"net/http"

	"github.com/urandom/embed/filesystem"
)

// NewFileSystem creates a new empty filesystem.
func NewFileSystem() (http.FileSystem, error) {
	fs := filesystem.New()

	return fs, nil
}
