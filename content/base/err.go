package base

type Error struct {
	err error
}

func (e Error) Err() error {
	return e.err
}

func (e *Error) SetErr(err error) {
	e.err = err
}
