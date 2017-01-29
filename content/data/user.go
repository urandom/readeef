package data

import (
	"database/sql/driver"
	"encoding/gob"
	"fmt"
)

type Login string

type User struct {
	Login       Login
	FirstName   string `db:"first_name"`
	LastName    string `db:"last_name"`
	Email       string
	HashType    string `db:"hash_type" json:"-"`
	Admin       bool
	Active      bool
	ProfileJSON []byte `db:"profile_data" json:"-"`
	Salt        []byte `json:"-"`
	Hash        []byte `json:"-"`
	MD5API      []byte `db:"md5_api" json:"-"` // "md5(user:pass)"

	ProfileData map[string]interface{}
}

func (val *Login) Scan(src interface{}) error {
	switch t := src.(type) {
	case string:
		*val = Login(t)
	case []byte:
		*val = Login(t)
	default:
		return fmt.Errorf("Scan source '%#v' (%T) was not of type string (Login)", src, src)
	}

	return nil
}

func (val Login) Value() (driver.Value, error) {
	return string(val), nil
}

func init() {
	gob.Register(Login(""))
}
