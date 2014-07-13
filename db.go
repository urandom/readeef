package readeef

import (
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const (
	exists_user = `SELECT 1 FROM users WHERE login = ?;`

	get_user    = `SELECT first_name, last_name, email, salt, hash, md5_api FROM users WHERE login = ?;`
	create_user = `
INSERT INTO users(first_name, last_name, email, salt, hash, md5_api, login)
	VALUES(?, ?, ?, ?, ?, ?, ?);`
	update_user = `
UPDATE users SET first_name = ?, last_name = ?, email = ?, salt = ?, hash = ?, md5_api = ?
	WHERE login = ?;`
)

var (
	init_sql = []string{`
CREATE TABLE IF NOT EXISTS users (
	login TEXT PRIMARY KEY,
	first_name TEXT,
	last_name TEXT,
	email TEXT,
	salt TEXT,
	hash TEXT,
	md5_api TEXT
);`,
	}
)

type Validator interface {
	Validate() error
}

type DB struct {
	*sqlx.DB
	driver        string
	connectString string
}

type ValidationError error

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

func (db DB) GetUser(login string) (User, error) {
	var u User
	if err := db.Get(&u, get_user, login); err != nil {
		return User{}, err
	}

	u.Login = login

	return u, nil
}

func (db DB) UpdateUser(u User) error {
	if err := u.Validate(); err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	row := db.QueryRow(exists_user, u.Login)
	var exists int
	row.Scan(&exists)

	var stmt *sqlx.Stmt

	if exists == 1 {
		stmt, err = tx.Preparex(update_user)
	} else {
		stmt, err = tx.Preparex(create_user)
	}

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.FirstName, u.LastName, u.Email, u.Salt, u.Hash, u.MD5API, u.Login)
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func (db DB) init() error {
	for _, sql := range init_sql {
		_, err := db.Exec(sql)
		if err != nil {
			return errors.New(fmt.Sprintf("Error executing '%s': %v", sql, err))
		}
	}

	return nil
}
