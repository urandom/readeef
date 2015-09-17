package base

import (
	"reflect"

	"github.com/jmoiron/sqlx"
	"github.com/urandom/readeef/content/sql/db"
)

type Helper struct {
	sql db.SqlStmts
}

func NewHelper() *Helper {
	return &Helper{sql: sqlStmts}
}

func (h Helper) SQL() db.SqlStmts {
	return h.sql
}

func (h *Helper) Set(override db.SqlStmts) {
	oursPtr := reflect.ValueOf(&h.sql)
	ours := oursPtr.Elem()
	theirs := reflect.ValueOf(override)

	for i := 0; i < ours.NumField(); i++ {
		ourInner := ours.Field(i)
		theirInner := theirs.Field(i)

		if theirInner.IsValid() {
			for j := 0; j < theirInner.NumField(); j++ {
				ourField := ourInner.Field(j)
				theirField := theirInner.Field(j)

				if theirField.IsValid() && ourField.CanSet() && ourField.Kind() == reflect.String {
					s := theirField.String()
					if s != "" {
						ourField.SetString(s)
					}
				}
			}
		}
	}
}

func (h Helper) Upgrade(db *db.DB, old, new int) error {
	return nil
}

func (h Helper) CreateWithId(tx *sqlx.Tx, sql string, args ...interface{}) (int64, error) {
	var id int64

	stmt, err := tx.Preparex(sql)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(args...)
	if err != nil {
		return 0, err
	}

	id, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

var (
	sqlStmts = db.SqlStmts{}
)
