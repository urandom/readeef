package pool

import (
	"bytes"
	"sync"
)

type bufp struct {
	sync.Pool
}

// Buffer is a utility variable that provides bytes.Buffer objects.
var Buffer bufp

// Get returns a bytes.Buffer pointer from the pool.
func (b *bufp) Get() *bytes.Buffer {
	buffer := b.Pool.Get().(*bytes.Buffer)
	buffer.Reset()

	return buffer
}

func init() {
	Buffer = bufp{
		Pool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}
