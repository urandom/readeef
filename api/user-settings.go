package api

import (
	"encoding/json"
	"net/http"
	"readeef"

	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
)

type UserSettings struct {
	webfw.BaseController
}

func NewUserSettings() UserSettings {
	return UserSettings{
		webfw.NewBaseController("/v:version/user/settings", webfw.MethodGet|webfw.MethodPost, ""),
	}
}

func (con UserSettings) Handler(c context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resp interface{}
		var err error

		db := readeef.GetDB(c)
		user := readeef.GetUser(c, r)

		if !user.Active {
			readeef.Debug.Println("User " + user.Login + " is inactive")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if r.Method == "GET" {
			type response struct {
				Login       string
				ProfileData map[string]interface{}
			}

			resp = response{
				Login:       user.Login,
				ProfileData: user.ProfileData,
			}
		} else if r.Method == "POST" {
			dec := json.NewDecoder(r.Body)

			err = dec.Decode(&user.ProfileData)
			if err == nil {
				err = db.UpdateUser(user)

				if err == nil {
					type response struct {
						Success bool
					}
					resp = response{true}
				}
			}

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

func (con UserSettings) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}
