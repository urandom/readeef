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
		webfw.NewBaseController("/v:version/user/:attribute", webfw.MethodGet|webfw.MethodPost, ""),
	}
}

func (con UserSettings) Handler(c context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		db := readeef.GetDB(c)
		user := readeef.GetUser(c, r)

		if !user.Active {
			readeef.Debug.Println("User " + user.Login + " is inactive")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		params := webfw.GetParams(c, r)
		attr := params["attribute"]

		resp := map[string]interface{}{"Login": user.Login}
		if r.Method == "GET" {
			switch attr {
			case "firstName":
				resp[attr] = user.FirstName
			case "lastName":
				resp[attr] = user.LastName
			case "email":
				resp[attr] = user.Email
			case "profileData":
				resp[attr] = user.ProfileData
			default:
				err = errors.New("Error getting user attribute: unknown attribute " + attr)
			}
		} else if r.Method == "POST" {
			buf := util.BufferPool.GetBuffer()
			defer util.BufferPool.Put(buf)

			buf.ReadFrom(r.Body)

			switch attr {
			case "firstName":
				user.FirstName = buf.String()
			case "lastName":
				user.LastName = buf.String()
			case "email":
				user.Email = buf.String()
			case "password":
				err = user.SetPassword(buf.String())
			case "profileData":
				err = json.Unmarshal(buf.Bytes(), &user.ProfileData)
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
