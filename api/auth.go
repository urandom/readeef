package api

import (
	"net/http"

	"github.com/urandom/handler/auth"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

func tokenCreate(repo content.Repo, secret []byte, log readeef.Logger) http.Handler {
	return auth.TokenGenerator(nil, auth.AuthenticatorFunc(func(user, pass string) bool {
		u := repo.UserByLogin(data.Login(user))
		if u.HasErr() {
			log.Infof("Error fetching user %s: %+v", user, u.Err())
		}
		return u.Authenticate(pass, secret)
	}), secret, auth.Logger(log))
}

func tokenDelete(storage content.TokenStorage, secret []byte, log readeef.Logger) http.Handler {
	return auth.TokenBlacklister(nil, storage, secret, auth.Logger(log))
}
