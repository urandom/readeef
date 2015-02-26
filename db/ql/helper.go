package ql

import (
	"github.com/urandom/readeef/db"
	"github.com/urandom/readeef/db/base"
)

type Helper struct {
	base.Helper
}

func (h Helper) Init() []string {
	return initSQL
}

func init() {
	helper := &Helper{}

	db.Register("ql", helper)
}
