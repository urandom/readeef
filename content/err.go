package content

import "github.com/pkg/errors"

var (
	ErrNoContent = errors.New("No content")
)

func IsNoContent(err error) bool {
	cause := errors.Cause(err)

	return cause == ErrNoContent
}
