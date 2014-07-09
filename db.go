package readeef

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var db *sqlx.DB

const (
	create_user_table = `
CREATE TABLE IF NOT EXISTS users (
	login TEXT PRIMARY KEY
	first_name TEXT
	last_name TEXT
	email TEXT
	salt TEXT
	hash TEXT
	md5api TEXT
);`

	exists_user = `SELECT 1 FROM users WHERE login = ?;`

	get_user    = `SELECT first_name, last_name, email, salt, hash, md5api FROM users WHERE login = ?;`
	create_user = `
INSERT INTO users(first_name, last_name, email, salt, hash, md5api, login)
	VALUES(?, ?, ?, ?, ?, ?, ?);`
	update_user = `
UPDATE users SET first_name = ?, last_name = ?, email = ?, salt = ?, hash = ?, md5api = ? 
	WHERE login = ?;`
)

func getUser(login string) (User, error) {
	var u User
	if err := db.Get(&u, get_user, login); err != nil {
		return User{}, err
	}

	u.Login = login

	return u, nil
}

func updateUser(u User) error {
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

func init() {
	var err error

	db, err = sqlx.Connect("sqlite3", "./readeef.sqlite")
	if err != nil {
		log.Fatalln(err)
	}

	db.MustExec(create_user_table)
}
