package repo

import (
	"fmt"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/sql/postgres"
	"github.com/urandom/readeef/content/sql/sqlite3"
	"github.com/urandom/readeef/db"
	"github.com/urandom/webfw"
)

func New(db *db.DB, logger webfw.Logger) content.Repo {
	switch db.DriverName() {
	case "postgres":
		return postgres.NewRepo(db, logger)
	case "sqlite3":
		return sqlite3.NewRepo(db, logger)
	default:
		panic(fmt.Sprintf("Cannot provide a repo for driver '%s'\n", db.DriverName()))
	}
}
