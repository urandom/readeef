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
	getUser         = `SELECT first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api FROM users WHERE login = :login`
	getUserByMD5Api = `SELECT login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash FROM users WHERE md5_api = :md5_api`
	getUsers        = `SELECT login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api FROM users`

	createUser = `
INSERT INTO users(login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api)
	SELECT :login, :first_name, :last_name, :email, :admin, :active, :profile_data, :hash_type, :salt, :hash, :md5_api EXCEPT
	SELECT login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api FROM users WHERE login = :login`
	updateUser = `
UPDATE users SET first_name = :first_name, last_name = :last_name, email = :email, admin = :admin, active = :active, profile_data = :profile_data, hash_type = :hash_type, salt = :salt, hash = :hash, md5_api = :md5_api
	WHERE login = :login`
	deleteUser = `DELETE FROM users WHERE login = :login`
)
