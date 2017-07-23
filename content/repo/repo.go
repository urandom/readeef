package repo

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/sql/db"
	"github.com/urandom/readeef/content/sql/postgres"
	"github.com/urandom/readeef/content/sql/sqlite3"
)

func New(driver, source string, log readeef.Logger) (r content.Repo, err error) {
	switch driver {
	case "sqlite3", "postgres":
		db := db.New(log)
		if err = db.Open(driver, source); err != nil {
			err = errors.Wrap(err, "connecting to database")
			return
		}

		switch driver {
		case "postgres":
			r = postgres.NewRepo(db, log)
		case "sqlite3":
			r = sqlite3.NewRepo(db, log)
		}

		return
	default:
		panic(fmt.Sprintf("Cannot provide a repo for driver '%s'\n", driver))
	}
}
