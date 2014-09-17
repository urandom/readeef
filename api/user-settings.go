package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"readeef"

	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/util"
)

type UserSettings struct {
	webfw.BaseController
}

func NewUserSettings() UserSettings {
	return UserSettings{
		webfw.NewBaseController("/v:version/user-settings/:attribute", webfw.MethodGet|webfw.MethodPost, ""),
	}
}

func (con UserSettings) Handler(c context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		db := readeef.GetDB(c)
		user := readeef.GetUser(c, r)

		params := webfw.GetParams(c, r)
		attr := params["attribute"]

		resp := map[string]interface{}{"Login": user.Login}
		if r.Method == "GET" {
			switch attr {
			case "FirstName":
				resp[attr] = user.FirstName
			case "LastName":
				resp[attr] = user.LastName
			case "Email":
				resp[attr] = user.Email
			case "ProfileData":
				resp[attr] = user.ProfileData
			default:
				err = errors.New("Error getting user attribute: unknown attribute " + attr)
			}
		} else if r.Method == "POST" {
			buf := util.BufferPool.GetBuffer()
			defer util.BufferPool.Put(buf)

			buf.ReadFrom(r.Body)

			switch attr {
			case "FirstName":
				user.FirstName = buf.String()
			case "LastName":
				user.LastName = buf.String()
			case "Email":
				user.Email = buf.String()
			case "ProfileData":
				err = json.Unmarshal(buf.Bytes(), &user.ProfileData)
			case "password":
				data := struct {
					Current string
					New     string
				}{}
				err = json.Unmarshal(buf.Bytes(), &data)
				if err == nil {
					/* TODO: non-fatal error */
					if user.Authenticate(data.Current) {
						err = user.SetPassword(data.New)
					} else {
						err = errors.New("Error change user password: current password is invalid")
					}
				}
			default:
				err = errors.New("Error getting user attribute: unknown attribute " + attr)
			}

			if err == nil {
				err = db.UpdateUser(user)
			}

			if err == nil {
				resp["Success"] = true
				resp["Attribute"] = attr
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
