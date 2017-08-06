package api

import (
	"net/http"

	"github.com/go-chi/chi"
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

func listUsers(repo repo.User, log log.Log) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		users, err := repo.All()
		if err != nil {
			errors(w, log, "Error getting users: %+v", err)
			return
		}

		args{"users": users}.WriteJSON(w)
	}
}

type addUserData struct {
	Login     content.Login `json:"login"`
	FirstName string        `json:"firstName"`
	LastName  string        `json:"lastName"`
	Email     string        `json:"email"`
	Admin     bool          `json:"admin"`
	Active    bool          `json:"active"`
	Password  string        `json:"password"`
}

func addUser(repo repo.User, secret []byte, log log.Log) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		var userData addUserData
		if stop := readJSON(w, r.Body, &userData); stop {
			return
		}

		u, err := repo.Get(userData.Login)
		if err == nil {
			http.Error(w, "User exists", http.StatusConflict)
			return
		} else if !content.IsNoContent(err) {
			error(w, log, "Error getting user: %+v", err)
			return
		}

		u = content.User{
			Login:     userData.Login,
			FirstName: userData.FirstName,
			LastName:  userData.LastName,
			Email:     userData.Email,
			Admin:     userData.Admin,
			Active:    userData.Active,
		}

		if err = u.Password(userData.Password, secret); err != nil {
			error(w, log, "Error setting user password: %+v", err)
			return
		}

		if err = repo.Update(u); err != nil {
			error(w, log, "Error updating user: %+v", err)
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
			error(w, log, "Error deleting user: %+v", err)
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
