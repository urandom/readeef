package readeef

import (
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

var init_sql = make(map[string][]string)
var sql_stmt = make(map[string]string)

type Validator interface {
	Validate() error
}

type DB struct {
	*sqlx.DB
	driver        string
	connectString string
}

type ValidationError struct {
	error
}

func NewDB(driver, conn string) DB {
	return DB{driver: driver, connectString: conn}
}

func (db *DB) Connect() error {
	dbx, err := sqlx.Connect(db.driver, db.connectString)
	if err != nil {
		return err
	}

	db.DB = dbx

	return db.init()
}

func (db DB) init() error {
	if init, ok := init_sql[db.driver]; ok {
		for _, sql := range init {
			_, err := db.Exec(sql)
			if err != nil {
				return errors.New(fmt.Sprintf("Error executing '%s': %v", sql, err))
			}
		}
	} else {
		return errors.New(fmt.Sprintf("No init sql for driver '%s'", db.driver))
	}

	return nil
}

func (db DB) NamedSQL(name string) string {
	var stmt string

	if stmt = sql_stmt[db.driver+":"+name]; stmt == "" {
		stmt = sql_stmt["generic:"+name]
	}

	return stmt
}

func (db DB) CreateWithId(stmt *sqlx.Stmt, args ...interface{}) (int64, error) {
	var id int64

	if db.driver == "postgres" {
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
