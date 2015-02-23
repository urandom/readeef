package base

import "github.com/urandom/readeef/content"

type Error struct {
	err error
}

func (e Error) Err() error {
	return e.err
}

func (e *Error) SetErr(err error) content.Error {
	e.err = err

	return e
}
