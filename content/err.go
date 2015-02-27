package content

type Error interface {
	Err(err ...error) error
	HasErr() bool
}
