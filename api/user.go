package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/urandom/readeef"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/util"
)

type User struct{}

func NewUser() User {
	return User{}
}

func (con User) Patterns() map[string]webfw.MethodIdentifierTuple {
	prefix := "/v:version/user/"

	return map[string]webfw.MethodIdentifierTuple{
		prefix + "list":                 webfw.MethodIdentifierTuple{webfw.MethodGet, "list"},
		prefix + "add/:login":           webfw.MethodIdentifierTuple{webfw.MethodPost, "add"},
		prefix + "remove/:login":        webfw.MethodIdentifierTuple{webfw.MethodPost, "remove"},
		prefix + "active/:login/:state": webfw.MethodIdentifierTuple{webfw.MethodPost, "active"},
	}
}

func (con User) Handler(c context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error

		db := readeef.GetDB(c)
		user := readeef.GetUser(c, r)

		if !user.Admin {
			readeef.Debug.Println("User " + user.Login + " is not an admin")

			w.WriteHeader(http.StatusForbidden)
			return
		}

		action := webfw.GetMultiPatternIdentifier(c, r)
		params := webfw.GetParams(c, r)
		resp := make(map[string]interface{})

		switch action {
		case "list":
			users, err := db.GetUsers()
			if err != nil {
				break
			}

			type user struct {
				Login     string
				FirstName string
				LastName  string
				Email     string
				Active    bool
				Admin     bool
			}

			userList := []user{}
			for _, u := range users {
				userList = append(userList, user{
					Login:     u.Login,
					FirstName: u.FirstName,
					LastName:  u.LastName,
					Email:     u.Email,
					Active:    u.Active,
					Admin:     u.Admin,
				})
			}

			resp["Users"] = userList
		case "add":
			login := params["login"]

			_, err = db.GetUser(login)
			/* TODO: non-fatal error */
			if err == nil {
				err = errors.New("User with login " + login + " already exists")
				break
			} else if err != sql.ErrNoRows {
				break
			}

			buf := util.BufferPool.GetBuffer()
			defer util.BufferPool.Put(buf)

			buf.ReadFrom(r.Body)

			u := readeef.User{Login: login}

			err = u.SetPassword(buf.String())
			if err != nil {
				break
			}

			err = db.UpdateUser(u)
			if err != nil {
				break
			}

			resp["Success"] = true
			resp["Login"] = login
		case "remove":
			login := params["login"]

			if user.Login == login {
				err = errors.New("The current user cannot be removed")
				break
			}

			var u readeef.User

			u, err = db.GetUser(login)
			if err != nil {
				break
			}

			err = db.DeleteUser(u)
			if err != nil {
				break
			}

			resp["Success"] = true
			resp["Login"] = login
		case "active":
			login := params["login"]

			if user.Login == login {
				err = errors.New("The current user cannot be removed")
				break
			}

			active := params["state"] == "true"

			var u readeef.User

			u, err = db.GetUser(login)
			if err != nil {
				break
			}

			u.Active = active
			err = db.UpdateUser(u)
			if err != nil {
				break
			}

			resp["Success"] = true
			resp["Login"] = login
		}

		var b []byte
		if err == nil {
			b, err = json.Marshal(resp)
		}
		if err != nil {
			webfw.GetLogger(c).Print(err)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(b)
	})
}

func (con User) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}
