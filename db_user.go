package readeef

import "encoding/json"

const (
	get_user    = `SELECT first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api FROM users WHERE login = $1`
	create_user = `
INSERT INTO users(login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api)
	SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11 EXCEPT
	SELECT login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api FROM users WHERE login = $1`
	update_user = `
UPDATE users SET first_name = $1, last_name = $2, email = $3, admin = $4, active = $5, profile_data = $6, hash_type = $7, salt = $8, hash = $9, md5_api = $10
	WHERE login = $11`
	delete_user = `DELETE FROM users WHERE login = $1`

	get_users = `SELECT login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api FROM users`

	get_feed_users = `
SELECT u.login, u.first_name, u.last_name, u.email, u.admin, u.active,
	   u.profile_data, u.hash_type, u.salt, u.hash, u.md5_api
FROM users u, users_feeds uf
WHERE u.login = uf.user_login AND uf.feed_id = $1
`
)

func (db DB) GetUser(login string) (User, error) {
	var u User
	if err := db.Get(&u, db.NamedSQL("get_user"), login); err != nil {
		return u, err
	}

	u.Login = login
	users, err := initUsers([]User{u})
	if err != nil {
		return u, err
	}

	return users[0], nil
}

func (db DB) GetUsers() ([]User, error) {
	var users []User
	if err := db.Select(&users, db.NamedSQL("get_users")); err != nil {
		return users, err
	}

	return initUsers(users)
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

	res, err := ustmt.Exec(u.FirstName, u.LastName, u.Email, u.Admin, u.Active, u.ProfileJSON, u.HashType, u.Salt, u.Hash, u.MD5API, u.Login)
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

	_, err = cstmt.Exec(u.Login, u.FirstName, u.LastName, u.Email, u.Admin, u.Active, u.ProfileJSON, u.HashType, u.Salt, u.Hash, u.MD5API)
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

func (db DB) GetFeedUsers(f Feed) ([]User, error) {
	var users []User
	if err := db.Select(&users, db.NamedSQL("get_feed_users"), f.Id); err != nil {
		return users, err
	}

	return initUsers(users)
}

func initUsers(users []User) ([]User, error) {
	for i := 0; i < len(users); i++ {
		u := &users[i]
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

func init() {
	sql_stmt["generic:get_user"] = get_user
	sql_stmt["generic:create_user"] = create_user
	sql_stmt["generic:update_user"] = update_user
	sql_stmt["generic:delete_user"] = delete_user
	sql_stmt["generic:get_users"] = get_users
	sql_stmt["generic:get_feed_users"] = get_feed_users
}
