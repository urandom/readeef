package api

import (
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/urandom/handler/auth"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
)

func tokenCreate(repo repo.User, secret []byte, log readeef.Logger) http.Handler {
	return auth.TokenGenerator(nil, auth.AuthenticatorFunc(func(user, pass string) bool {
		u := repo.Get(content.Login(user))
		if u.HasErr() {
			log.Infof("Error fetching user %s: %+v", user, u.Err())
		}
		return u.Authenticate(pass, secret)
	}), secret, auth.Logger(log))
}

func tokenDelete(storage content.TokenStorage, secret []byte, log readeef.Logger) http.Handler {
	return auth.TokenBlacklister(nil, storage, secret, auth.Logger(log))
}

func tokenValidator(
	repo repo.User,
	storage content.TokenStorage,
	log readeef.Logger,
) auth.TokenValidator {
	return auth.TokenValidatorFunc(func(token string, claims jwt.Claims) bool {
		exists, err := storage.Exists(token)

		if err != nil {
			log.Printf("Error using token storage: %+v\n", err)
			return false
		}

		if exists {
			return false
		}

		if c, ok := claims.(*jwt.StandardClaims); ok {
			u := repo.Get(content.Login(c.Subject))
			err := u.Err()

			if err != nil {
				if err != content.ErrNoContent {
					log.Printf("Error getting user %s from repo: %+v\n", c.Subject, err)
				}

				return false
			}
		}

		return true
	})
}
