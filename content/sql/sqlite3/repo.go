package sqlite3

import (
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content/sql"
	"github.com/urandom/readeef/content/sql/db"
)

type Repo struct {
	*sql.Repo
}

func NewRepo(db *db.DB, log readeef.Logger) *Repo {
	return &Repo{Repo: sql.NewRepo(db, log)}
}
