package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"readeef"
	"strings"

	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/util"
)

type User struct {
	webfw.BaseController
}

func NewUser() User {
	return User{
		webfw.NewBaseController("/v:version/user/*action", webfw.MethodGet|webfw.MethodPost, ""),
	}
}

func (con User) Handler(c context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		db := readeef.GetDB(c)
		user := readeef.GetUser(c, r)

		if !user.Active || !user.Admin {
			if !user.Active {
				readeef.Debug.Println("User " + user.Login + " is inactive")
			} else {
				readeef.Debug.Println("User " + user.Login + " is not an admin")
			}

			w.WriteHeader(http.StatusForbidden)
			return
		}

		actionParam := webfw.GetParams(c, r)
		parts := strings.Split(actionParam["action"], "/")
		action := parts[0]

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
			if len(parts) != 2 {
				err = errors.New(fmt.Sprintf("Expected 2 arguments, got %d", len(parts)))
				break
			}

			login := parts[1]

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
			if len(parts) != 2 {
				err = errors.New(fmt.Sprintf("Expected 2 arguments, got %d", len(parts)))
				break
			}

			login := parts[1]

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
			if len(parts) != 3 {
				err = errors.New(fmt.Sprintf("Expected 3 arguments, got %d", len(parts)))
				break
			}

			login := parts[1]

			if user.Login == login {
				err = errors.New("The current user cannot be removed")
				break
			}

			active := parts[2] == "true"

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
		default:
			err = errors.New("Error processing request: unknown action " + action)
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
	}
}

func (con User) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}
