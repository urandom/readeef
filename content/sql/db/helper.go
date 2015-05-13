package db

import "github.com/jmoiron/sqlx"

type Helper interface {
	SQL(name string) string
	InitSQL() []string

	CreateWithId(tx *sqlx.Tx, name string, args ...interface{}) (int64, error)
	Upgrade(db *DB, old, new int) error
}

func Register(driver string, helper Helper) {
	if helper == nil {
		panic("No helper provided")
	}

	if _, ok := helpers[driver]; ok {
		panic("Helper " + driver + " already registered")
	}

	helpers[driver] = helper
}

// Can't recover from missing driver or statement, panic
func (db DB) SQL(name string) string {
	driver := db.DriverName()

	if h, ok := helpers[driver]; ok {
		sql := h.SQL(name)

		if sql == "" {
			panic("No statement registered under " + name)
		}

		return sql
	} else {
		panic("No helper registered for " + driver)
	}
}
