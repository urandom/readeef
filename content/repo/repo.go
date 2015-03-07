package repo

import (
	"fmt"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/sql/postgres"
	"github.com/urandom/readeef/content/sql/sqlite3"
	"github.com/urandom/readeef/content/sql/db"
	"github.com/urandom/webfw"
)

func New(driver, source string, logger webfw.Logger) (r content.Repo, err error) {
	switch driver {
	case "sqlite3", "postgres":
		db := db.New(logger)
		if err = db.Open(driver, source); err != nil {
			err = fmt.Errorf("Error connecting to database: %v", err)
			return
		}

		switch driver {
		case "postgres":
			r = postgres.NewRepo(db, logger)
		case "sqlite3":
			r = sqlite3.NewRepo(db, logger)
		}

		return
	default:
		panic(fmt.Sprintf("Cannot provide a repo for driver '%s'\n", driver))
	}
}
