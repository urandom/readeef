package api

import (
	"context"
	"encoding/json"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/urandom/handler/auth"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type contextKey string

var userKey = contextKey("user")

func userContext(repo repo.User, next http.Handler, log log.Log) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c, ok := auth.Claims(r).(*jwt.StandardClaims); ok {
			user, err := repo.Get(content.Login(c.Subject))

			if err != nil {
				if content.IsNoContent(err) {
					http.Error(w, "Not found", http.StatusNotFound)
					return
				}

				fatal(w, log, "Error loading user: %+v", err)
				return
			}

			ctx := context.WithValue(r.Context(), userKey, user)

			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			http.Error(w, "Invalid claims", http.StatusBadRequest)
		}
	})
}

func userValidator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, stop := userFromRequest(w, r); stop {
			return
		}

		next.ServeHTTP(w, r)

	})
}

func userFromRequest(w http.ResponseWriter, r *http.Request) (user content.User, stop bool) {
	var ok bool
	if user, ok = r.Context().Value(userKey).(content.User); ok {
		return user, false
	}

	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	return content.User{}, true
}

type args map[string]interface{}

func (a args) WriteJSON(w http.ResponseWriter) {
	b, err := json.Marshal(a)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Write(b)
}
