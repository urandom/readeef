package api

import (
	"context"
	"encoding/json"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/urandom/handler/auth"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
)

func userContext(repo repo.User, next http.Handler, log readeef.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c, ok := auth.Claims(r).(*jwt.StandardClaims); ok {
			user := repo.Get(content.Login(c.Subject))

			if user.HasErr() {
				err := user.Err()
				if err == content.ErrNoContent {
					http.Error(w, "Not found", http.StatusNotFound)
					return
				} else {
					log.Printf("Error loading user %s: %+v", c.Subject, err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			ctx := context.WithValue(r.Context(), "user", user)

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
		} else {
			next.ServeHTTP(w, r)
		}

	})
}

func userFromRequest(w http.ResponseWriter, r *http.Request) (user content.User, stop bool) {
	var ok bool
	if user, ok = r.Context().Value("user").(content.User); ok {
		return user, false
	}

	http.Error(w, "Bad Request", http.StatusBadRequest)
	return nil, true
}

type args map[string]interface{}

func (a args) WriteJSON(w http.ResponseWriter) {
	b, err := json.Marshal(a)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Write(b)
}
