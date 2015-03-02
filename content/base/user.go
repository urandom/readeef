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
	"github.com/urandom/readeef/content/data"
)

type User struct {
	ArticleSorting
	Error
	RepoRelated
	ArticleSearch

	data data.User
}

type UserRelated struct {
	user content.User
}

func (u User) String() string {
	if u.data.Email == "" {
		return u.data.FirstName + " " + u.data.LastName
	} else {
		return u.data.FirstName + " " + u.data.LastName + " <" + u.data.Email + ">"
	}
}

func (u *User) Data(d ...data.User) data.User {
	if u.HasErr() {
		return u.data
	}

	if len(d) > 0 {
		data := d[0]
		var err error

		if len(data.ProfileJSON) == 0 {
			data.ProfileJSON, err = json.Marshal(data.ProfileData)
		} else {
			if err = json.Unmarshal(data.ProfileJSON, &data.ProfileData); err != nil {
				u.Err(err)
				return u.data
			}

			if data.ProfileData == nil {
				data.ProfileData = make(map[string]interface{})
			}
		}

		u.Err(err)
		u.data = data
	}

	return u.data
}

func (u User) Validate() error {
	if u.data.Login == "" {
		return ValidationError{errors.New("Invalid user login")}
	}
	if u.data.Email != "" {
		if _, err := mail.ParseAddress(u.String()); err != nil {
			return ValidationError{err}
		}
	}

	return nil
}

func (u User) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.data)
}

func (u *User) Password(password string, secret []byte) {
	if u.HasErr() {
		return
	}

	h := md5.Sum([]byte(fmt.Sprintf("%s:%s", u.data.Login, password)))

	u.data.MD5API = h[:]

	c := 30
	salt := make([]byte, c)
	if _, err := rand.Read(salt); err != nil {
		u.Err(err)
		return
	}

	u.data.Salt = salt

	u.data.HashType = "sha1"
	u.data.Hash = u.generateHash(password, secret)
}

func (u User) Authenticate(password string, secret []byte) bool {
	if u.HasErr() {
		return false
	}

	return bytes.Equal(u.data.Hash, u.generateHash(password, secret))
}

func (u User) generateHash(password string, secret []byte) []byte {
	hash := sha1.Sum(append(secret, append(u.data.Salt, []byte(password)...)...))

	return hash[:]
}

func (ur *UserRelated) User(us ...content.User) content.User {
	if len(us) > 0 {
		ur.user = us[0]
	}

	return ur.user
}
