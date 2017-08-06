package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/log"
)

type DB struct {
	*sqlx.DB
	log log.Log
}

var (
	dbVersion = 4

	helpers = make(map[string]Helper)
)

func New(log log.Log) *DB {
	return &DB{log: log}
}

func (db *DB) Open(driver, connect string) (err error) {
	db.DB, err = sqlx.Connect(driver, connect)

	if err == nil {
		err = db.init()
	}

	return
}

func (db *DB) CreateWithID(tx *sqlx.Tx, sql string, args ...interface{}) (int64, error) {
	driver := db.DriverName()

	if h, ok := helpers[driver]; ok {
		return h.CreateWithID(tx, sql, args...)
	} else {
		panic("No helper registered for " + driver)
	}
}

func (db *DB) WhereMultipleORs(column string, length, off int) string {
	if length < 20 {
		orSlice := make([]string, length)
		for i := 0; i < length; i++ {
			orSlice[i] = fmt.Sprintf("%s = $%d", column, off+i)
		}

		return "(" + strings.Join(orSlice, " OR ") + ")"
	}

	driver := db.DriverName()
	if h, ok := helpers[driver]; ok {
		return h.WhereMultipleORs(column, length, off)
	} else {
		panic("No helper registered for " + driver)
	}
}

func (db *DB) init() error {
	helper := helpers[db.DriverName()]

	if helper == nil {
		return errors.Errorf("no helper provided for driver '%s'", db.DriverName())
	}

	for _, sql := range helper.InitSQL() {
		_, err := db.Exec(sql)
		if err != nil {
			return errors.Wrapf(err, "executing '%s'", sql)
		}
	}

	var version int
	if err := db.Get(&version, "SELECT db_version FROM readeef"); err != nil {
		if err == sql.ErrNoRows {
			version = dbVersion
		} else {
			return errors.Wrap(err, "getting the current db_version")
		}
	}

	if version > dbVersion {
		panic(fmt.Sprintf("The db version '%d' is newer than the expected '%d'", version, dbVersion))
	}

	if version < dbVersion {
		db.log.Infof("Database version mismatch: current is %d, expected %d\n", version, dbVersion)
		db.log.Infof("Running upgrade function for %s driver\n", db.DriverName())
		if err := helper.Upgrade(db, version, dbVersion); err != nil {
			return errors.Wrapf(err, "Error running upgrade function for %s driver", db.DriverName())
		}
	}

	_, err := db.Exec(`DELETE FROM readeef`)
	/* TODO: per-database statements */
	if err == nil {
		_, err = db.Exec(`INSERT INTO readeef(db_version) VALUES($1)`, dbVersion)
	}
	if err != nil {
		return errors.Wrap(err, "initializing readeef utility table")
	}

	return nil
}
