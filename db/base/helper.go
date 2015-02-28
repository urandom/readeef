package base

import "github.com/urandom/readeef/db"

type Helper struct {
}

func (h Helper) SQL(name string) string {
	return sql[name]
}

func (h Helper) Set(name, stmt string) {
	sql[name] = stmt
}

func (h Helper) Upgrade(db *db.DB, old, new int) error {
	return nil
}

func (h Helper) CreateWithId(tx *db.Tx, name string, args ...interface{}) (int64, error) {
	var id int64

	sql := h.SQL(name)
	if sql == "" {
		panic("No statement registered under " + name)
	}

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
	sql = make(map[string]string)
)
