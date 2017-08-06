package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
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

	var val interface{}
	switch chi.URLParam(r, "key") {
	case firstNameSetting:
		val = user.FirstName
	case lastNameSetting:
		val = user.LastName
	case emailSetting:
		val = user.Email
	case profileSetting:
		val = user.ProfileData
	case activeSetting:
		val = user.Active
	default:
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	args{"value": val}.WriteJSON(w)
}

func setSettingValue(repo repo.User, secret []byte, log log.Log) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		if name := content.Login(chi.URLParam(r, "name")); name != "" {
			var err error
			user, err = repo.Get(name)
			if err != nil {
				error(w, log, "Error getting user: %+v", err)
				return
			}
		}

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			error(w, log, "Error reading request body: %+v", err)
			return
		}

		switch chi.URLParam(r, "key") {
		case firstNameSetting:
			err = json.Unmarshal(b, &user.FirstName)
		case lastNameSetting:
			err = json.Unmarshal(b, &user.LastName)
		case emailSetting:
			err = json.Unmarshal(b, &user.Email)
		case profileSetting:
			if err = json.Unmarshal(b, &user.ProfileData); err == nil {
				user.ProfileJSON = b
			}
		case activeSetting:
			user.Active = string(b) == "true"
		case passwordSetting:
			passwd := struct {
				Current string
				New     string
			}{}

			if err = json.Unmarshal(b, &passwd); err == nil {
				var auth bool
				if auth, err = user.Authenticate(passwd.Current, secret); auth {
					err = user.Password(passwd.New, secret)
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
			error(w, log, "Error setting user setting: %+v", err)
		}
	}
}
