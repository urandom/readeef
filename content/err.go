package content

type Error interface {
	Err() error
	SetErr(err error)
}
