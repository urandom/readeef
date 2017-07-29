// +build nofs

package readeef

import "net/http"

// NewFileSystem creates a new http.Dir filesystem
func NewFileSystem() (http.FileSystem, error) {
	return http.Dir("."), nil
}
