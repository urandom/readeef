package content

import "errors"

type Error interface {
	Err(err ...error) error
	HasErr() bool
}

var (
	ErrNoContent = errors.New("No content")
)

func IsNoContent(err error) bool {
	cause := errors.Cause(err)

	return cause == ErrNoContent
}
