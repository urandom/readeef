package api

import (
	"encoding/json"
	"errors"
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

		if !user.Active {
			readeef.Debug.Println("User " + user.Login + " is inactive")
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
		case "create":
			buf := util.BufferPool.GetBuffer()
			defer util.BufferPool.Put(buf)

			buf.ReadFrom(r.Body)

			data := struct {
				Login    string
				Password string
			}{}

			err = json.Unmarshal(buf.Bytes(), &data)
			if err != nil {
				break
			}

			u := readeef.User{Login: data.Login}

			err = u.SetPassword(data.Password)
			if err != nil {
				break
			}

			err = db.UpdateUser(u)
			if err != nil {
				break
			}

			resp["Success"] = true
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
