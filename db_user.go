package readeef

import "encoding/json"

const (
	get_user    = `SELECT first_name, last_name, email, profile_data, hash_type, salt, hash, md5_api FROM users WHERE login = $1`
	create_user = `
INSERT INTO users(login, first_name, last_name, email, profile_data, hash_type, salt, hash, md5_api)
	SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9 EXCEPT
	SELECT login, first_name, last_name, email, profile_data, hash_type, salt, hash, md5_api FROM users WHERE login = $1`
	update_user = `
UPDATE users SET first_name = $1, last_name = $2, email = $3, profile_data = $4, hash_type = $5, salt = $6, hash = $7, md5_api = $8
	WHERE login = $9`
	delete_user = `DELETE FROM users WHERE login = $1`

	get_users = `SELECT login, first_name, last_name, email, profile_data, hash_type, salt, hash, md5_api FROM users`
)

func (db DB) GetUser(login string) (User, error) {
	var u User
	if err := db.Get(&u, db.NamedSQL("get_user"), login); err != nil {
		return u, err
	}

	u.Login = login
	if len(u.ProfileJSON) != 0 {
		if err := json.Unmarshal(u.ProfileJSON, &u.ProfileData); err != nil {
			return u, err
		}
	}
	if u.ProfileData == nil {
		u.ProfileData = make(map[string]interface{})
	}

	return u, nil
}

func (db DB) GetUsers() ([]User, error) {
	var users []User
	if err := db.Select(&users, db.NamedSQL("get_users")); err != nil {
		return users, err
	}

	for _, u := range users {
		if len(u.ProfileJSON) != 0 {
			if err := json.Unmarshal(u.ProfileJSON, &u.ProfileData); err != nil {
				return users, err
			}
		}
		if u.ProfileData == nil {
			u.ProfileData = make(map[string]interface{})
		}
	}

	return users, nil
}

func (db DB) UpdateUser(u User) error {
	var err error
	if err = u.Validate(); err != nil {
		return err
	}

	u.ProfileJSON, err = json.Marshal(u.ProfileData)
	if err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ustmt, err := tx.Preparex(db.NamedSQL("update_user"))

	if err != nil {
		return err
	}
	defer ustmt.Close()

	res, err := ustmt.Exec(u.FirstName, u.LastName, u.Email, u.ProfileJSON, u.HashType, u.Salt, u.Hash, u.MD5API, u.Login)
	if err != nil {
		return err
	}

	if num, err := res.RowsAffected(); err == nil && num > 0 {
		tx.Commit()
		return nil
	}

	cstmt, err := tx.Preparex(db.NamedSQL("create_user"))

	if err != nil {
		return err
	}
	defer cstmt.Close()

	_, err = cstmt.Exec(u.Login, u.FirstName, u.LastName, u.Email, u.ProfileJSON, u.HashType, u.Salt, u.Hash, u.MD5API)
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
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.NamedSQL("delete_user"))

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

func init() {
	sql_stmt["generic:get_user"] = get_user
	sql_stmt["generic:create_user"] = create_user
	sql_stmt["generic:update_user"] = update_user
	sql_stmt["generic:delete_user"] = delete_user
	sql_stmt["generic:get_users"] = get_users
}
