package base

type Error struct {
	err error
}

func (e *Error) Err(err ...error) error {
	prev := e.err

	if len(err) > 0 {
		e.err = err[0]
	} else {
		e.err = nil
	}

	return prev
}

func (e Error) HasErr() bool {
	return e.err != nil
}
