package base

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/info"
)

type User struct {
	ArticleSorting
	Error
	RepoRelated

	info info.User
}

type UserRelated struct {
	user content.User
}

func (u User) String() string {
	if u.info.Email == "" {
		return u.info.FirstName + " " + u.info.LastName
	} else {
		return u.info.FirstName + " " + u.info.LastName + " <" + u.info.Email + ">"
	}
}

func (u *User) Info(in ...info.User) info.User {
	if u.Err() != nil {
		return u.info
	}

	if len(in) > 0 {
		info := in[0]
		var err error

		if len(info.ProfileJSON) == 0 {
			info.ProfileJSON, err = json.Marshal(info.ProfileData)
		} else {
			if len(info.ProfileJSON) != 0 {
				if err = json.Unmarshal(info.ProfileJSON, &info.ProfileData); err != nil {
					u.Err(err)
					return u.info
				}
			}
			if info.ProfileData == nil {
				info.ProfileData = make(map[string]interface{})
			}
		}

		u.Err(err)
		u.info = info
	}

	return u.info
}

func (u User) Validate() error {
	if u.info.Login == "" {
		return ValidationError{errors.New("Invalid user login")}
	}
	if u.info.Email != "" {
		if _, err := mail.ParseAddress(u.String()); err != nil {
			return ValidationError{err}
		}
	}

	return nil
}

func (u *User) Password(password string, secret []byte) {
	if u.Err() != nil {
		return
	}

	h := md5.Sum([]byte(fmt.Sprintf("%s:%s", u.info.Login, password)))

	u.info.MD5API = h[:]

	c := 30
	salt := make([]byte, c)
	if _, err := rand.Read(salt); err != nil {
		u.Err(err)
		return
	}

	u.info.Salt = salt

	u.info.HashType = "sha1"
	u.info.Hash = u.generateHash(password, secret)
}

func (u User) Authenticate(password string, secret []byte) bool {
	return bytes.Equal(u.info.Hash, u.generateHash(password, secret))
}

func (u User) generateHash(password string, secret []byte) []byte {
	hash := sha1.Sum(append(secret, append(u.info.Salt, []byte(password)...)...))

	return hash[:]
}

func (ur *UserRelated) User(us ...content.User) content.User {
	if len(us) > 0 {
		ur.user = us[0]
	}

	return ur.user
}
