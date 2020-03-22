package db

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cenkalti/backoff/v3"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/log"
)

type DB struct {
	*sqlx.DB

	stmtCache      map[string]*sqlx.Stmt
	namedStmtCache map[string]*sqlx.NamedStmt

	log log.Log

	mu     sync.RWMutex
	helper Helper
}

var (
	dbVersion = 4

	helpers = make(map[string]Helper)
)

func New(log log.Log) *DB {
	return &DB{stmtCache: make(map[string]*sqlx.Stmt), namedStmtCache: make(map[string]*sqlx.NamedStmt), log: log}
}

func (db *DB) Open(driver, connect string) (err error) {
	if u, err := url.Parse(connect); err == nil && u.Scheme == "file" {
		if dir := filepath.Dir(u.Opaque); dir != "" {
			if err := os.MkdirAll(dir, 0700); err != nil {
				return errors.Wrapf(err, "creating db directory %s", dir)
			}
		}
	}
	db.DB, err = sqlx.Connect(driver, connect)

	if err == nil {
		err = db.init()
	}

	return err
}

func (db *DB) CreateWithID(tx *sqlx.Tx, sql string, arg interface{}) (id int64, err error) {
	err = backoff.Retry(func() error {
		id, err = db.helper.CreateWithID(tx, sql, arg)
		return err
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5))

	return id, err
}

func (db *DB) WhereMultipleORs(column, prefix string, length int, equal bool) string {
	if length < 20 {
		orSlice := make([]string, length)
		sign := "="
		if !equal {
			sign = "!="
		}
		for i := 0; i < length; i++ {
			orSlice[i] = fmt.Sprintf("%s %s :%s%d", column, sign, prefix, i)
		}

		return "(" + strings.Join(orSlice, " OR ") + ")"
	}

	return db.helper.WhereMultipleORs(column, prefix, length, equal)
}

func (db *DB) WithNamedStmt(query string, tx *sqlx.Tx, cb func(*sqlx.NamedStmt) error) error {
	stmt, err := db.getNamedStmt(query)
	if err != nil {
		return errors.WithMessage(err, "preparing named statement")
	}

	if tx != nil {
		stmt = tx.NamedStmt(stmt)
	}

	return backoff.Retry(func() error {
		err := cb(stmt)
		if err == nil {
			return nil
		}
		if !db.helper.RetryableErr(err) {
			return backoff.Permanent(err)
		}
		db.log.Debugln("Retrying query", query)
		return err
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5))
}

func (db *DB) WithStmt(query string, tx *sqlx.Tx, cb func(*sqlx.Stmt) error) error {
	stmt, err := db.getStmt(query)
	if err != nil {
		return errors.WithMessage(err, "preparing statement")
	}

	if tx != nil {
		stmt = tx.Stmtx(stmt)
	}

	return backoff.Retry(func() error {
		err := cb(stmt)
		if err == nil {
			return nil
		}
		if !db.helper.RetryableErr(err) {
			return backoff.Permanent(err)
		}
		db.log.Debugln("Retrying query", query)
		return err
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5))
}

func (db *DB) WithTx(cb func(*sqlx.Tx) error) error {
	tx, err := db.Beginx()
	if err != nil {
		return errors.WithMessage(err, "creating transaction")
	}
	defer tx.Rollback()

	if err := cb(tx); err != nil {
		return errors.WithMessage(err, "executing transaction")
	}

	if err := tx.Commit(); err != nil {
		return errors.WithMessage(err, "committing transaction")
	}

	return nil
}

func (db *DB) init() error {
	helper := helpers[db.DriverName()]

	if helper == nil {
		return errors.Errorf("no helper provided for driver '%s'", db.DriverName())
	}
	db.helper = helper

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

func (db *DB) getNamedStmt(query string) (stmt *sqlx.NamedStmt, err error) {
	db.mu.RLock()
	stmt = db.namedStmtCache[query]
	db.mu.RUnlock()

	if stmt == nil {
		if stmt, err = db.PrepareNamed(query); err == nil {
			db.mu.Lock()
			db.namedStmtCache[query] = stmt
			db.mu.Unlock()
		}
	}

	return stmt, err
}

func (db *DB) getStmt(query string) (stmt *sqlx.Stmt, err error) {
	if stmt = db.stmtCache[query]; stmt == nil {
		if stmt, err = db.Preparex(query); err == nil {
			db.stmtCache[query] = stmt
		}
	}

	return stmt, err
}
