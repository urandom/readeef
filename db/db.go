package db

import (
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/urandom/webfw"
)

type Helper interface {
	SQL(name string, sql ...string) string
	Init() []string

	Upgrade(db DB, old, new int) error
}

type DB struct {
	*sqlx.DB
	logger webfw.Logger
	frozen *Tx
}

var (
	errAlreadyFrozen = errors.New("DB already frozen")
	errNotFrozen     = errors.New("DB not frozen")

	helpers = make(map[string]Helper)
)

func New(logger webfw.Logger) *DB {
	return &DB{logger: logger}
}

func Register(name string, helper Helper) {
	if helper == nil {
		panic("No helper provided")
	}

	if _, ok := helpers[name]; ok {
		panic("Helper " + name + " already registered")
	}

	helpers[name] = helper
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

func (db *DB) CreateWithId(tx *Tx, sql string, args ...interface{}) (int64, error) {
	var id int64

	stmt, err := tx.Preparex(sql)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	if db.DriverName() == "postgres" {
		sql += " RETURNING id"
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
