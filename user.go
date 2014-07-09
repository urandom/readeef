package readeef

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
)

type User struct {
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Email     string
	Login     string
	Salt      []byte
	Hash      []byte
	MD5API    []byte `db:"md5_api"` // "md5(user:pass)"

	config Config
}

func (u User) setPassword(pass string) (User, error) {
	h := md5.Sum([]byte(fmt.Sprintf("%s:%s", u.Login, pass)))

	u.MD5API = h[:]

	c := 30
	salt := make([]byte, c)
	if _, err := rand.Read(salt); err != nil {
		return u, err
	}

	u.Salt = salt

	u.Hash = u.generateHash(pass)

	return u, nil
}

func (u User) Authenticate(pass string) bool {
	return bytes.Equal(u.Hash, u.generateHash(pass))
}

func (u User) generateHash(pass string) []byte {
	hash := sha1.Sum(append([]byte(u.config.Auth.Secret), append(u.Salt, []byte(pass)...)...))

	return hash[:]
}
