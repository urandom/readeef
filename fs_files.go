// +build !nofs

// DO NOT EDIT ** This file was generated with github.com/urandom/embed ** DO NOT EDIT //

package readeef

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/urandom/embed/filesystem"
)

// NewFileSystem creates a new filesystem with pre-filled binary data.
func NewFileSystem() (http.FileSystem, error) {
	fs := filesystem.New()

	fs.Fallback = true

	if err := fs.Add("rf-ng/ui/index.html", 231, os.FileMode(420), time.Unix(1506035621, 0), "<!doctype html>\n<html lang=\"en\">\n<head>\n  <meta charset=\"utf-8\">\n  <title>readeef: feed aggregator</title>\n  <script>\n    location.href = \"/\" + localStorage.getItem(\"locale\") || \"en\"+ \"/\"\n  </script>\n</head>\n<body>\n</body>\n</html>\n"); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("packing file rf-ng/ui/index.html"))
	}

	return fs, nil
}
