package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/util"
)

type UserSettings struct {
	webfw.BasePatternController
}

type getUserAttributeProcessor struct {
	Attribute string `json:"attribute"`

	user content.User
}

type setUserAttributeProcessor struct {
	Attribute string          `json:"attribute"`
	Value     json.RawMessage `json:"value"`

	user   content.User
	secret []byte
}

func NewUserSettings() UserSettings {
	return UserSettings{
		webfw.NewBasePatternController("/v:version/user-settings/:attribute", webfw.MethodGet|webfw.MethodPost, ""),
	}
}

func (con UserSettings) Handler(c context.Context) http.Handler {
	cfg := readeef.GetConfig(c)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := readeef.GetUser(c, r)

		params := webfw.GetParams(c, r)
		attr := params["attribute"]

		var resp responseError
		if r.Method == "GET" {
			resp = getUserAttribute(user, attr)
		} else if r.Method == "POST" {
			buf := util.BufferPool.GetBuffer()
			defer util.BufferPool.Put(buf)

			buf.ReadFrom(r.Body)

			resp = setUserAttribute(user, []byte(cfg.Auth.Secret), attr, buf.Bytes())
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

func (p getUserAttributeProcessor) Process() responseError {
	return getUserAttribute(p.user, p.Attribute)
}

func (p setUserAttributeProcessor) Process() responseError {
	return setUserAttribute(p.user, p.secret, p.Attribute, p.Value)
}

func getUserAttribute(user content.User, attr string) (resp responseError) {
	resp = newResponse()
	in := user.Data()
	resp.val["Login"] = in.Login

	switch attr {
	case "FirstName":
		resp.val[attr] = in.FirstName
	case "LastName":
		resp.val[attr] = in.LastName
	case "Email":
		resp.val[attr] = in.Email
	case "ProfileData":
		resp.val[attr] = in.ProfileData
	default:
		resp.err = errors.New("Error getting user attribute: unknown attribute " + attr)
		return
	}

	return
}

func setUserAttribute(user content.User, secret []byte, attr string, data []byte) (resp responseError) {
	resp = newResponse()
	in := user.Data()
	resp.val["Login"] = in.Login

	switch attr {
	case "FirstName":
		resp.err = json.Unmarshal(data, &in.FirstName)
	case "LastName":
		resp.err = json.Unmarshal(data, &in.LastName)
	case "Email":
		resp.err = json.Unmarshal(data, &in.Email)
	case "ProfileData":
		if resp.err = json.Unmarshal(data, &in.ProfileData); resp.err == nil {
			in.ProfileJSON = []byte{}
		}
	case "Active":
		in.Active = string(data) == "true"
	case "Password":
		passwd := struct {
			Current string
			New     string
		}{}
		if resp.err = json.Unmarshal(data, &passwd); resp.err != nil {
			/* TODO: non-fatal error */
			return
		}
		if user.Authenticate(passwd.Current, secret) {
			user.Password(passwd.New, secret)
			resp.err = user.Err()
		} else {
			resp.err = errors.New("Error change user password: current password is invalid")
		}
	default:
		resp.err = errors.New("Error getting user attribute: unknown attribute " + attr)
	}

	if resp.err != nil {
		return
	}

	user.Data(in)
	user.Update()
	if resp.err = user.Err(); resp.err != nil {
		return
	}

	resp.val["Success"] = true
	resp.val["Attribute"] = attr

	return
}
