package base

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/tests"
)

func TestUser(t *testing.T) {
	u := User{}
	u.data.FirstName = "First"
	u.data.LastName = "Last"

	tests.CheckString(t, "First Last", u.String())

	u.data.Email = "example@sugr.org"
	tests.CheckString(t, "First Last <example@sugr.org>", u.String())

	d := u.Data()

	tests.CheckString(t, "example@sugr.org", d.Email)

	d = u.Data(data.User{Email: ""})

	tests.CheckString(t, "", d.Email)
	tests.CheckString(t, "", d.FirstName)

	tests.CheckBool(t, false, u.Validate() == nil)

	u.data.Email = "example"
	tests.CheckBool(t, false, u.Validate() == nil)

	u.data.Email = "example@sugr.org"
	tests.CheckBool(t, false, u.Validate() == nil)

	u.data.Login = data.Login("login")
	tests.CheckBool(t, true, u.Validate() == nil)

	ejson, eerr := json.Marshal(u.data)
	tests.CheckBool(t, true, eerr == nil)

	ajson, aerr := json.Marshal(u)
	tests.CheckBool(t, true, aerr == nil)

	tests.CheckBytes(t, ejson, ajson)

	tests.CheckBytes(t, []byte{}, u.data.MD5API)
	tests.CheckString(t, "", u.data.HashType)
	tests.CheckBytes(t, []byte{}, u.data.Hash)

	u.Password("pass", []byte("secret"))
	h := md5.Sum([]byte(fmt.Sprintf("%s:%s", "login", "pass")))
	tests.CheckBytes(t, h[:], u.data.MD5API)
	tests.CheckString(t, "sha1", u.data.HashType)
	tests.CheckBytes(t, u.generateHash("pass", []byte("secret")), u.data.Hash)
}
