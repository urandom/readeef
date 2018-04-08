package api

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/urandom/handler/auth"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

func getUserData(w http.ResponseWriter, r *http.Request) {
	user, stop := userFromRequest(w, r)
	if stop {
		return
	}

	args{"user": user}.WriteJSON(w)
}

func createUserToken(secret []byte, log log.Log) http.HandlerFunc {
	generator := auth.TokenGenerator(nil, auth.AuthenticatorFunc(func(user, pass string) bool {
		// We're already logged in at this point, and the data is fake
		return true
	}), secret, auth.Logger(log))

	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		r.Form.Set("user", string(user.Login))
		r.Form.Set("password", "phony")

		generator.ServeHTTP(w, r)
	}
}

func listUsers(repo repo.User, log log.Log) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, stop := userFromRequest(w, r)
		if stop {
			return
		}

		users, err := repo.All()
		if err != nil {
			fatal(w, log, "Error getting users: %+v", err)
			return
		}

		args{"users": users}.WriteJSON(w)
	}
}

func addUser(repo repo.User, secret []byte, log log.Log) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, stop := userFromRequest(w, r)
		if stop {
			return
		}

		login := content.Login(r.Form.Get("login"))

		_, err := repo.Get(login)
		if err == nil {
			http.Error(w, "User exists", http.StatusConflict)
			return
		} else if !content.IsNoContent(err) {
			fatal(w, log, "Error getting user: %+v", err)
			return
		}

		u := content.User{
			Login:     login,
			FirstName: r.Form.Get("firstName"),
			LastName:  r.Form.Get("lastName"),
			Email:     r.Form.Get("email"),
		}

		if _, ok := r.Form["admin"]; ok {
			u.Admin = true
		}

		if _, ok := r.Form["active"]; ok {
			u.Active = true
		}

		if err = u.Password(r.Form.Get("password"), secret); err != nil {
			fatal(w, log, "Error setting user password: %+v", err)
			return
		}

		if err = repo.Update(u); err != nil {
			fatal(w, log, "Error updating user: %+v", err)
			return
		}

		args{"success": true}.WriteJSON(w)
	}
}

func deleteUser(repo repo.User, log log.Log) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		name := content.Login(chi.URLParam(r, "name"))
		if user.Login == name {
			http.Error(w, "Current user", http.StatusConflict)
			return
		}

		u, err := repo.Get(name)
		if err == nil {
			err = repo.Delete(u)
		}

		if err != nil {
			fatal(w, log, "Error deleting user: %+v", err)
			return
		}

		args{"success": true}.WriteJSON(w)
	}
}

func adminValidator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		if !user.Admin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)

	})
}
