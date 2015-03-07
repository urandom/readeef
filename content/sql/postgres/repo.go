package postgres

import (
	"github.com/urandom/readeef/content/sql"
	"github.com/urandom/readeef/content/sql/db"
	"github.com/urandom/webfw"
)

type Repo struct {
	*sql.Repo
}

func NewRepo(db *db.DB, logger webfw.Logger) *Repo {
	return &Repo{Repo: sql.NewRepo(db, logger)}
}
