package dgraph

import (
	"github.com/dgraph-io/dgo"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/log"
)

type articleRepo struct {
	dg *dgo.Dgraph

	log log.Log
}

func (r articleRepo) ForUser(user content.User, opts ...content.QueryOpt) ([]content.Article, error) {
	panic("not implemented")
}

func (r articleRepo) All(opts ...content.QueryOpt) ([]content.Article, error) {
	panic("not implemented")
}

func (r articleRepo) Count(user content.User, opts ...content.QueryOpt) (int64, error) {
	panic("not implemented")
}

func (r articleRepo) IDs(user content.User, opts ...content.QueryOpt) ([]content.ArticleID, error) {
	panic("not implemented")
}

func (r articleRepo) Read(state bool, user content.User, opts ...content.QueryOpt) error {
	panic("not implemented")
}

func (r articleRepo) Favor(state bool, user content.User, opts ...content.QueryOpt) error {
	panic("not implemented")
}

func (r articleRepo) RemoveStaleUnreadRecords() error {
	panic("not implemented")
}
