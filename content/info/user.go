package info

type Login string

type User struct {
	Login       Login
	FirstName   string `db:"first_name"`
	LastName    string `db:"last_name"`
	Email       string
	HashType    string `db:"hash_type"`
	Admin       bool
	Active      bool
	ProfileJSON []byte `db:"profile_data"`
	Salt        []byte
	Hash        []byte
	MD5API      []byte `db:"md5_api"` // "md5(user:pass)"

	ProfileData map[string]interface{}
}
