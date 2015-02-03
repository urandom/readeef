package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/urandom/readeef"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/util"
)

type UserSettings struct {
	webfw.BasePatternController
}

func NewUserSettings() UserSettings {
	return UserSettings{
		webfw.NewBasePatternController("/v:version/user-settings/:attribute", webfw.MethodGet|webfw.MethodPost, ""),
	}
}

func (con UserSettings) Handler(c context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db := readeef.GetDB(c)
		user := readeef.GetUser(c, r)

		params := webfw.GetParams(c, r)
		attr := params["attribute"]

		var resp responseError
		if r.Method == "GET" {
			resp = getUserAttribute(db, user, attr)
		} else if r.Method == "POST" {
			buf := util.BufferPool.GetBuffer()
			defer util.BufferPool.Put(buf)

			buf.ReadFrom(r.Body)

			resp = setUserAttribute(db, user, attr, buf.String())
		}

		var b []byte
		if resp.err == nil {
			b, resp.err = json.Marshal(resp.val)
		}

		if resp.err != nil {
			webfw.GetLogger(c).Print(resp.err)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(b)
	})
}

func (con UserSettings) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}

func getUserAttribute(db readeef.DB, user readeef.User, attr string) (resp responseError) {
	resp = newResponse()
	resp.val["Login"] = user.Login

	switch attr {
	case "FirstName":
		resp.val[attr] = user.FirstName
	case "LastName":
		resp.val[attr] = user.LastName
	case "Email":
		resp.val[attr] = user.Email
	case "ProfileData":
		resp.val[attr] = user.ProfileData
	default:
		resp.err = errors.New("Error getting user attribute: unknown attribute " + attr)
		return
	}

	return
}

func setUserAttribute(db readeef.DB, user readeef.User, attr string, data string) (resp responseError) {
	resp = newResponse()
	resp.val["Login"] = user.Login

	switch attr {
	case "FirstName":
		user.FirstName = data
	case "LastName":
		user.LastName = data
	case "Email":
		user.Email = data
	case "ProfileData":
		resp.err = json.Unmarshal([]byte(data), &user.ProfileData)
	case "Active":
		resp.err = json.Unmarshal([]byte(data), &user.Active)
	case "password":
		passwd := struct {
			Current string
			New     string
		}{}
		if resp.err = json.Unmarshal([]byte(data), &passwd); resp.err != nil {
			/* TODO: non-fatal error */
			return
		}
		if user.Authenticate(passwd.Current) {
			resp.err = user.SetPassword(passwd.New)
		} else {
			resp.err = errors.New("Error change user password: current password is invalid")
		}
	default:
		resp.err = errors.New("Error getting user attribute: unknown attribute " + attr)
	}

	if resp.err != nil {
		return
	}

	if resp.err = db.UpdateUser(user); resp.err != nil {
		return
	}

	resp.val["Success"] = true
	resp.val["Attribute"] = attr

	return
}
