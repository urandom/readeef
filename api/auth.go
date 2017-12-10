package api

import (
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/urandom/handler/auth"
	"github.com/urandom/readeef/api/token"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

func tokenCreate(repo repo.User, secret []byte, log log.Log) http.Handler {
	return auth.TokenGenerator(nil, auth.AuthenticatorFunc(func(user, pass string) bool {
		u, err := repo.Get(content.Login(user))
		if err != nil {
			log.Infof("Error fetching user %s: %+v", user, err)
			return false
		}

		ok, err := u.Authenticate(pass, secret)
		if err != nil {
			log.Infof("Error authenticating user %s: %+v", user, err)
			return false
		}
		return ok
	}), secret, auth.Logger(log))
}

func tokenDelete(storage token.Storage, secret []byte, log log.Log) http.Handler {
	return auth.TokenBlacklister(nil, storage, secret, auth.Logger(log))
}

func tokenValidator(
	repo repo.User,
	storage token.Storage,
	log log.Log,
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
			_, err := repo.Get(content.Login(c.Subject))

			if err != nil {
				if !content.IsNoContent(err) {
					log.Printf("Error getting user %s from repo: %+v\n", c.Subject, err)
				}

				return false
			}

			return true
		}

		return false
	})
}
