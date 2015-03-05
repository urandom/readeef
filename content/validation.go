package content

type ValidationError struct {
	error
}

func NewValidationError(err error) ValidationError {
	return ValidationError{err}
}
