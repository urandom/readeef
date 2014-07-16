package readeef

import "github.com/jmoiron/sqlx"

const (
	exists_user = `SELECT 1 FROM users WHERE login = ?;`

	get_user    = `SELECT first_name, last_name, email, salt, hash, md5_api FROM users WHERE login = ?;`
	create_user = `
INSERT INTO users(first_name, last_name, email, salt, hash, md5_api, login)
	VALUES(?, ?, ?, ?, ?, ?, ?);`
	update_user = `
UPDATE users SET first_name = ?, last_name = ?, email = ?, salt = ?, hash = ?, md5_api = ?
	WHERE login = ?;`
	delete_user = `DELETE FROM users WHERE login = ?`
)

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

func (db DB) DeleteUser(u User) error {
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

	if exists != 1 {
		return nil
	}
	stmt, err = tx.Preparex(delete_user)

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Login)
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}
