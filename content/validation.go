package content

import "github.com/pkg/errors"

type ValidationError struct {
	error
}

func NewValidationError(err error) error {
	return errors.WithStack(ValidationError{err})
}
