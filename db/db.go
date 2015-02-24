package db

import (
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/urandom/webfw"
)

type DB struct {
	*sqlx.DB
	logger webfw.Logger
	frozen *Tx
}

var errAlreadyFrozen = errors.New("DB already frozen")
var errNotFrozen = errors.New("DB not frozen")

func New(logger webfw.Logger) *DB {
	return &DB{logger: logger}
}

func (db *DB) Open(driver, connect string) (err error) {
	db.DB, err = sqlx.Connect(driver, connect)

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

func (db *DB) CreateWithId(stmt *sqlx.Stmt, args ...interface{}) (int64, error) {
	var id int64

	if db.DriverName() == "postgres" {
		err := stmt.QueryRow(args...).Scan(&id)
		if err != nil {
			return 0, err
		}
	} else {
		res, err := stmt.Exec(args...)
		if err != nil {
			return 0, err
		}

		id, err = res.LastInsertId()
		if err != nil {
			return 0, err
		}
	}

	return id, nil
}
