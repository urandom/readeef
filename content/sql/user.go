package sql

import (
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/db"
	"github.com/urandom/webfw"
)

type User struct {
	base.User
	NamedSQL
	logger webfw.Logger

	db *db.DB
}

func NewUser(db *db.DB, logger webfw.Logger, authSecret []byte) User {
	u := User{NamedSQL: NewNamedSQL(), db: db, logger: logger}

	u.init()

	return u
}

func (u *User) Update() {
	if u.Err() != nil {
		return
	}

	i := u.Info()
	u.logger.Infof("Updating user %s\n", i.Login)

	tx, err := u.db.Begin()
	if err != nil {
		u.SetErr(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(u.SQL("update_user"))
	if err != nil {
		u.SetErr(err)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(i.FirstName, i.LastName, i.Email, i.Admin, i.Active, i.ProfileJSON, i.HashType, i.Salt, i.Hash, i.MD5API, i.Login)
	if err != nil {
		u.SetErr(err)
		return
	}

	if num, err := res.RowsAffected(); err == nil && num > 0 {
		tx.Commit()
		return
	}

	stmt, err = tx.Preparex(u.SQL("create_user"))
	if err != nil {
		u.SetErr(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(i.Login, i.FirstName, i.LastName, i.Email, i.Admin, i.Active, i.ProfileJSON, i.HashType, i.Salt, i.Hash, i.MD5API)
	if err != nil {
		u.SetErr(err)
		return
	}

	tx.Commit()

	return
}

func (u *User) Delete() {
	if u.Err() != nil {
		return
	}

	i := u.Info()
	u.logger.Infof("Deleting user %s\n", i.Login)

	if err := u.Validate(); err != nil {
		u.SetErr(err)
		return
	}

	tx, err := u.db.Begin()
	if err != nil {
		u.SetErr(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(u.SQL("delete_user"))
	if err != nil {
		u.SetErr(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(i.Login)
	if err != nil {
		u.SetErr(err)
		return
	}

	tx.Commit()
}

func (u *User) init() {
	u.SetSQL("create_user", createUser)
	u.SetSQL("update_user", updateUser)
	u.SetSQL("delete_user", deleteUser)
}

const (
	createUser = `
INSERT INTO users(login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api)
	SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11 EXCEPT
	SELECT login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api FROM users WHERE login = $1`
	updateUser = `
UPDATE users SET first_name = $1, last_name = $2, email = $3, admin = $4, active = $5, profile_data = $6, hash_type = $7, salt = $8, hash = $9, md5_api = $10
	WHERE login = $11`
	deleteUser = `DELETE FROM users WHERE login = $1`
)
