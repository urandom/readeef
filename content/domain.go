package content

import (
	"fmt"
	"net/url"
)

type Domain interface {
	Error
	RepoRelated

	fmt.Stringer

	URL(url ...string) *url.URL

	Validate() error

	SupportsHTTPS() bool
}
