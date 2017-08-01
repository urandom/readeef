package content

import "github.com/pkg/errors"

type ValidationError struct {
	error
}

func NewValidationError(err error) ValidationError {
	return ValidationError{errors.WithStack(err)}
}
