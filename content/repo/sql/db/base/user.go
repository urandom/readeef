package base

func init() {
	sqlStmts.User.Get = getUser
	sqlStmts.User.GetByMD5API = getUserByMD5Api
	sqlStmts.User.All = getUsers
	sqlStmts.User.Create = createUser
	sqlStmts.User.Update = updateUser
	sqlStmts.User.Delete = deleteUser
}

const (
	getUser         = `SELECT first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api FROM users WHERE login = $1`
	getUserByMD5Api = `SELECT login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash FROM users WHERE md5_api = $1`
	getUsers        = `SELECT login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api FROM users`

	createUser = `
INSERT INTO users(login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api)
	SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11 EXCEPT
	SELECT login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api FROM users WHERE login = $1`
	updateUser = `
UPDATE users SET first_name = $1, last_name = $2, email = $3, admin = $4, active = $5, profile_data = $6, hash_type = $7, salt = $8, hash = $9, md5_api = $10
	WHERE login = $11`
	deleteUser = `DELETE FROM users WHERE login = $1`
)
