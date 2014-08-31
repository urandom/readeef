package readeef

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"fmt"
	"net/mail"
)

type User struct {
	Login       string
	FirstName   string `db:"first_name"`
	LastName    string `db:"last_name"`
	Email       string
	HashType    string `db:"hash_type"`
	Admin       bool
	ProfileJSON []byte `db:"profile_data"`
	Salt        []byte
	Hash        []byte
	MD5API      []byte `db:"md5_api"` // "md5(user:pass)"

	ProfileData map[string]interface{}

	config Config
}

func (u User) String() string {
	if u.Email == "" {
		return u.FirstName + " " + u.LastName
	} else {
		return u.FirstName + " " + u.LastName + " <" + u.Email + ">"
	}
}

func (u *User) SetPassword(pass string) error {
	h := md5.Sum([]byte(fmt.Sprintf("%s:%s", u.Login, pass)))

	u.MD5API = h[:]

	c := 30
	salt := make([]byte, c)
	if _, err := rand.Read(salt); err != nil {
		return err
	}

	u.Salt = salt

	u.HashType = "sha1"
	u.Hash = u.generateHash(pass)

	return nil
}

func (u User) Authenticate(pass string) bool {
	return bytes.Equal(u.Hash, u.generateHash(pass))
}

func (u User) Validate() error {
	if u.Login == "" {
		return ValidationError{errors.New("Invalid user login")}
	}
	if u.Email != "" {
		if _, err := mail.ParseAddress(u.String()); err != nil {
			return ValidationError{err}
		}
	}

	return nil
}

func (u User) generateHash(pass string) []byte {
	hash := sha1.Sum(append([]byte(u.config.Auth.Secret), append(u.Salt, []byte(pass)...)...))

	return hash[:]
}
