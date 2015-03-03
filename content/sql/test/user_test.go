package test

import (
	"testing"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/tests"
)

func TestUser(t *testing.T) {
	u := repo.User()

	tests.CheckBool(t, false, u.HasErr(), u.Err())

	u.Update()
	tests.CheckBool(t, true, u.HasErr())

	err := u.Err()
	_, ok := err.(base.ValidationError)
	tests.CheckBool(t, true, ok, err)

	u.Data(data.User{Login: data.Login("login")})

	tests.CheckBool(t, false, u.HasErr(), u.Err())

	u.Update()
	tests.CheckBool(t, false, u.HasErr(), u.Err())

	u2 := repo.UserByLogin(data.Login("login"))
	tests.CheckBool(t, false, u2.HasErr(), u2.Err())
	tests.CheckString(t, "login", string(u2.Data().Login))

	u.Delete()
	tests.CheckBool(t, false, u.HasErr(), u.Err())

	u2 = repo.UserByLogin(data.Login("login"))
	tests.CheckBool(t, true, u2.HasErr())
	tests.CheckBool(t, true, u2.Err() == content.ErrNoContent)
}

func createUser(d data.User) (u content.User) {
	u = repo.User()
	u.Data(d)
	u.Update()

	return
}
