package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/urandom/webfw"
)

type DB struct {
	*sqlx.DB
	logger webfw.Logger
	frozen *Tx
}

var (
	dbVersion        = 1
	errAlreadyFrozen = errors.New("DB already frozen")
	errNotFrozen     = errors.New("DB not frozen")

	helpers = make(map[string]Helper)
)

func New(logger webfw.Logger) *DB {
	return &DB{logger: logger}
}

func (db *DB) Open(driver, connect string) (err error) {
	db.DB, err = sqlx.Connect(driver, connect)

	if err == nil {
		err = db.init()
	}

	return
}

func (db *DB) Freeze(tx *Tx) error {
	if db.frozen != nil {
		return errAlreadyFrozen
	}

	db.frozen = tx
	tx.frozen = true

	return nil
}

func (db *DB) Thaw() error {
	if db.frozen == nil {
		return errNotFrozen
	}

	db.frozen.frozen = false

	// Rollback in case the Tx wasn't completed successfully
	db.frozen.Rollback()

	return nil
}

func (db *DB) Begin() (*Tx, error) {
	if db.frozen != nil {
		return db.frozen, nil
	}

	if t, err := db.DB.Beginx(); err == nil {
		return &Tx{Tx: t}, nil
	} else {
		return nil, err
	}
}

func (db *DB) CreateWithId(tx *Tx, name string, args ...interface{}) (int64, error) {
	driver := db.DriverName()

	if h, ok := helpers[driver]; ok {
		return h.CreateWithId(tx, name, args)
	} else {
		panic("No helper registered for " + driver)
	}
}

func (db *DB) init() error {
	helper := helpers[db.DriverName()]

	if helper == nil {
		return fmt.Errorf("No helper provided for driver '%s'", db.DriverName())
	}

	for _, sql := range helper.Init() {
		_, err := db.Exec(sql)
		if err != nil {
			return fmt.Errorf("Error executing '%s': %v", sql, err)
		}
	}

	var version int
	if err := db.Get(&version, "SELECT db_version FROM readeef"); err != nil {
		if err == sql.ErrNoRows {
			version = dbVersion
		} else {
			return fmt.Errorf("Error getting the current db_version: %v\n", err)
		}
	}

	if version > dbVersion {
		panic(fmt.Sprintf("The db version '%d' is newer than the expected '%d'", version, dbVersion))
	}

	if version < dbVersion {
		db.logger.Infof("Database version mismatch: current is %d, expected %d\n", version, dbVersion)
		db.logger.Infof("Running upgrade function for %s driver\n", db.DriverName())
		if err := helper.Upgrade(db, version, dbVersion); err != nil {
			return fmt.Errorf("Error running upgrade function for %s driver: %v\n", db.DriverName(), err)
		}
	}

	_, err := db.Exec(`DELETE FROM readeef`)
	/* TODO: per-database statements */
	if err == nil {
		_, err = db.Exec(`INSERT INTO readeef(db_version) VALUES($1)`, dbVersion)
	}
	if err != nil {
		return fmt.Errorf("Error initializing readeef utility table: %v", err)
	}

	return nil
}
