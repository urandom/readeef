package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/urandom/readeef/content/data"
)

const (
	firstNameSetting = "first-name"
	lastNameSetting  = "last-name"
	emailSetting     = "email"
	profileSetting   = "profile"
	activeSetting    = "is-active"
	passwordSetting  = "password"
)

func getSettingKeys(w http.ResponseWriter, r *http.Request) {
	args{"keys": []string{
		firstNameSetting, lastNameSetting,
		emailSetting, profileSetting,
		activeSetting, passwordSetting,
	}}.WriteJSON(w)
}

func getSettingValue(w http.ResponseWriter, r *http.Request) {
	user, stop := userFromRequest(w, r)
	if stop {
		return
	}

	in := user.Data()

	var val interface{}
	switch chi.URLParam(r, "key") {
	case firstNameSetting:
		val = in.FirstName
	case lastNameSetting:
		val = in.LastName
	case emailSetting:
		val = in.Email
	case profileSetting:
		val = in.ProfileData
	case activeSetting:
		val = in.Active
	default:
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	args{"value": val}.WriteJSON(w)
}

func setSettingValue(secret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		if name := data.Login(chi.URLParam(r, "name")); name != "" {
			user = user.Repo().UserByLogin(name)
			if user.HasErr() {
				http.Error(w, "Error getting user: "+user.Err().Error(), http.StatusInternalServerError)
				return
			}
		}

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body: "+err.Error(), http.StatusInternalServerError)
			return
		}

		in := user.Data()

		switch chi.URLParam(r, "key") {
		case firstNameSetting:
			err = json.Unmarshal(b, &in.FirstName)
		case lastNameSetting:
			err = json.Unmarshal(b, &in.LastName)
		case emailSetting:
			err = json.Unmarshal(b, &in.Email)
		case profileSetting:
			if err = json.Unmarshal(b, &in.ProfileData); err == nil {
				in.ProfileJSON = []byte{}
			}
		case activeSetting:
			in.Active = string(b) == "true"
		case passwordSetting:
			passwd := struct {
				Current string
				New     string
			}{}

			if err = json.Unmarshal(b, &passwd); err == nil {
				if user.Authenticate(passwd.Current, secret) {
					user.Password(passwd.New, secret)
					err = user.Err()
				} else {
					http.Error(w, "Not authorized", http.StatusUnauthorized)
					return
				}
			}
		default:
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		if err == nil {
			args{"success": true}.WriteJSON(w)
		} else {
			http.Error(w, "Error parsing request body: "+err.Error(), http.StatusBadRequest)
		}
	}
}
