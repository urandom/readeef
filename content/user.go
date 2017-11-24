package content

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/mail"

	"golang.org/x/crypto/scrypt"

	"github.com/pkg/errors"
)

// Login is the user login name.
type Login string
type ProfileData map[string]interface{}

// User represents a readeef user.
type User struct {
	Login     Login  `json:"login"`
	FirstName string `db:"first_name" json:"firstName"`
	LastName  string `db:"last_name" json:"lastName"`
	Email     string `json:"email"`
	HashType  string `db:"hash_type" json:"-"`
	Admin     bool   `json:"admin"`
	Active    bool   `json:"active"`
	Salt      []byte `json:"-"`
	Hash      []byte `json:"-"`
	MD5API    []byte `db:"md5_api" json:"-"` // "md5(user:pass)"

	ProfileData ProfileData `db:"profile_data" json:"profileData"`
}

func (u *User) Password(password string, secret []byte) error {
	h := md5.Sum([]byte(fmt.Sprintf("%s:%s", u.Login, password)))

	u.MD5API = h[:]

	c := 30
	salt := make([]byte, c)
	if _, err := rand.Read(salt); err != nil {
		return errors.Wrap(err, "generating salt")
	}

	u.Salt = salt
	u.HashType = "scrypt"
	hash, err := u.generateHash(password, secret)
	if err != nil {
		return errors.WithMessage(err, "generating password hash")
	}

	u.Hash = hash

	return nil
}

// Validate checks whether all required fields have been provided.
func (u User) Validate() error {
	if u.Login == "" {
		return NewValidationError(errors.New("invalid user login"))
	}
	if u.Email != "" {
		if _, err := mail.ParseAddress(u.Email); err != nil {
			return NewValidationError(err)
		}
	}

	return nil
}

func (u User) Authenticate(password string, secret []byte) (bool, error) {
	hash, err := u.generateHash(password, secret)
	if err != nil {
		return false, errors.WithMessage(err, "authenticating user")
	}

	return subtle.ConstantTimeCompare(u.Hash, hash) == 1, nil
}

func (u User) String() string {
	if u.FirstName != "" && u.LastName != "" && u.Email != "" {
		return fmt.Sprintf("%s: %s %s (%s)", u.Login, u.FirstName, u.LastName, u.Email)
	} else if u.Email != "" {
		return fmt.Sprintf("%s: (%s)", u.Login, u.Email)
	} else {
		return string(u.Login)
	}
}

func (u User) generateHash(password string, secret []byte) ([]byte, error) {
	switch u.HashType {
	case "sha1":
		hash := sha1.Sum(append(secret, append(u.Salt, []byte(password)...)...))

		return hash[:], nil
	case "scrypt":
		dk, err := scrypt.Key([]byte(password), u.Salt, 16384, 8, 1, 32)
		if err != nil {
			err = errors.Wrap(err, "generating scrypt key")
		}

		return dk, err
	default:
		panic("Unknown hash type " + u.HashType)
	}
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

func (val *ProfileData) Scan(src interface{}) error {
	var data []byte
	switch t := src.(type) {
	case string:
		data = []byte(t)
	case []byte:
		data = t
	default:
		return fmt.Errorf("Scan source '%#v' (%T) was not of type string (Login)", src, src)
	}

	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, val)
}

func (val ProfileData) Value() (driver.Value, error) {
	return json.Marshal(val)
}

func (p *ProfileData) UnmarshalJSON(b []byte) error {
	data := map[string]json.RawMessage{}

	if err := json.Unmarshal(b, &data); err != nil {
		return errors.Wrap(err, "unmarshaling profile data")
	}

	var filters []Filter

	if len(data) > 0 {
		*p = ProfileData{}
	}

	for k, v := range data {
		switch k {
		case "filters":
			if err := json.Unmarshal(v, &filters); err != nil {
				return errors.Wrapf(err, "unmarshaling filters value %s", v)
			}
			(*p)[k] = filters
		default:
			var val interface{}

			if err := json.Unmarshal(v, &val); err != nil {
				return errors.Wrapf(err, "unmarshaling key %s value %s", k, v)
			}

			(*p)[k] = val
		}
	}

	return nil
}
