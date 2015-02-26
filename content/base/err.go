package base

type Error struct {
	err error
}

func (e Error) Err(err ...error) error {
	if len(err) > 0 {
		e.err = err[0]
	}
	return e.err
}
