package api

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

func getUserData(w http.ResponseWriter, r *http.Request) {
	user, stop := userFromRequest(w, r)
	if stop {
		return
	}

	args{"user": user}.WriteJSON(w)
}

func listUsers(w http.ResponseWriter, r *http.Request) {
	if user, stop := userFromRequest(w, r); stop {
		return
	} else {
		repo := user.Repo()

		users := repo.AllUsers()
		if repo.HasErr() {
			http.Error(w, "Error getting users: "+repo.Err().Error(), http.StatusInternalServerError)
			return
		}

		args{"users": users}.WriteJSON(w)
	}
}

type addUserData struct {
	Login     data.Login `json:"login"`
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	Email     string     `json:"email"`
	Admin     bool       `json:"admin"`
	Active    bool       `json:"active"`
	Password  string     `json:"password"`
}

func addUser(secret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		var userData addUserData
		if stop := readJSON(w, r.Body, &userData); stop {
			return
		}

		u := user.Repo().UserByLogin(userData.Login)
		if u.HasErr() {
			err := u.Err()
			if err != content.ErrNoContent {
				http.Error(w, "Error getting user: "+err.Error(), http.StatusInternalServerError)
				return
			}

			in := data.User{
				Login:     userData.Login,
				FirstName: userData.FirstName,
				LastName:  userData.LastName,
				Email:     userData.Email,
				Admin:     userData.Admin,
				Active:    userData.Active,
			}

			u := user.Repo().User()
			u.Data(in)
			u.Password(userData.Password, secret)

			u.Update()

			if u.HasErr() {
				http.Error(w, "Error creating user:"+u.Err().Error(), http.StatusInternalServerError)
			} else {
				args{"success": true}.WriteJSON(w)
			}
		} else {
			http.Error(w, "User exists", http.StatusConflict)
			return
		}
	}
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	user, stop := userFromRequest(w, r)
	if stop {
		return
	}

	name := data.Login(chi.URLParam(r, "name"))
	if user.Data().Login == name {
		http.Error(w, "Current user", http.StatusConflict)
		return
	}

	u := user.Repo().UserByLogin(name)
	u.Delete()

	if u.HasErr() {
		http.Error(w, "Error deleting user: "+u.Err().Error(), http.StatusInternalServerError)
	} else {
		args{"success": true}.WriteJSON(w)
	}
}

func adminValidator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user, stop := userFromRequest(w, r); stop {
			return
		} else {
			if !user.Data().Admin {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}

	})
}
