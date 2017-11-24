package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/mock_repo"
)

func Test_getSettingKeys(t *testing.T) {
	tests := []struct {
		name string
		keys []string
	}{
		{"all keys", []string{
			"first-name",
			"last-name",
			"email",
			"profile",
			"is-active",
			"password",
		}},
	}

	type data struct {
		Keys []string `json:"keys"`
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			getSettingKeys(w, r)

			if w.Code != http.StatusOK {
				t.Errorf("getSettingsKeys() code = %v, want %v", w.Code, http.StatusOK)
				return
			}

			got := data{}
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Errorf("getSettingKeys() body = '%s', error = %v", w.Body, err)
				return
			}

			if !reflect.DeepEqual(got.Keys, tt.keys) {
				t.Errorf("getUserData() user = %v, want = %v", got.Keys, tt.keys)
			}
		})
	}
}

func Test_getSettingValue(t *testing.T) {
	tests := []struct {
		name    string
		hasUser bool
		key     string
		value   interface{}
	}{
		{"no user", false, "", nil},
		{"unkown", true, "unknown key", nil},
		{"first name", true, firstNameSetting, "first"},
		{"last name", true, lastNameSetting, "last"},
		{"email", true, emailSetting, "email@example.com"},
		{"profile", true, profileSetting, content.ProfileData{"foo": float64(42)}},
		{"active", true, activeSetting, true},
	}

	type data struct {
		Value json.RawMessage `json:"value"`
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			code := http.StatusBadRequest
			var user content.User
			if tt.hasUser {
				user = content.User{Login: "test", FirstName: "first", LastName: "last", Email: "email@example.com", ProfileData: content.ProfileData{"foo": 42}, Active: true}
				r = r.WithContext(context.WithValue(r.Context(), userKey, user))
				r = addChiParam(r, "key", tt.key)

				switch tt.key {
				case firstNameSetting, lastNameSetting, emailSetting, profileSetting, activeSetting:
					code = http.StatusOK
				default:
					code = http.StatusNotFound
				}
			}

			getSettingValue(w, r)

			if w.Code != code {
				t.Errorf("getSettingsKeys() code = %v, want %v", w.Code, code)
				return
			}

			if code == http.StatusOK {
				got := data{}
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Errorf("getSettingKeys() body = '%s', error = %v", w.Body, err)
					return
				}

				var value interface{}
				switch tt.key {
				case firstNameSetting, lastNameSetting, emailSetting:
					var val string
					if err := json.Unmarshal(got.Value, &val); err != nil {
						t.Errorf("getSettingKeys() body = '%s', error = %v", got.Value, err)
						return
					}
					value = val
				case activeSetting:
					var val bool
					if err := json.Unmarshal(got.Value, &val); err != nil {
						t.Errorf("getSettingKeys() body = '%s', error = %v", got.Value, err)
						return
					}
					value = val
				case profileSetting:
					val := content.ProfileData{}
					if err := json.Unmarshal(got.Value, &val); err != nil {
						t.Errorf("getSettingKeys() body = '%s', error = %v", got.Value, err)
						return
					}
					value = val
				}

				if !reflect.DeepEqual(value, tt.value) {
					t.Errorf("getUserData() value = %v, want = %v", value, tt.value)
				}
			}
		})
	}
}

func Test_setSettingValue(t *testing.T) {
	tests := []struct {
		name         string
		hasUser      bool
		otherUser    bool
		otherUserErr bool
		key          string
		form         string
		formErr      bool
		updateErr    bool
	}{
		{"no user", false, false, false, "", "", false, false},
		{"other user error", true, true, true, "", "", false, false},
		{"first name", true, false, false, firstNameSetting, "value=first", false, false},
		{"last name", true, true, false, lastNameSetting, "value=last", false, false},
		{"email", true, true, false, emailSetting, "value=email@example.com", false, false},
		{"invalid email", true, true, false, emailSetting, "value=email", true, false},
		{"profile", true, true, false, profileSetting, `value={"foo": 42}`, false, false},
		{"active", true, true, false, activeSetting, `value=true`, false, false},
		{"password", true, false, false, passwordSetting, `value=newpass&current=pass`, false, false},
		{"unauthorized", true, false, false, passwordSetting, `value=newpass&current=pass1`, true, false},
		{"update error", true, false, false, firstNameSetting, `value=first`, false, true},
		{"unknown key", true, false, false, "unknown key", "", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userRepo := mock_repo.NewMockUser(ctrl)

			r := httptest.NewRequest("PUT", "/", strings.NewReader(tt.form))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()

			var code = http.StatusBadRequest
			var got content.User
			var u content.User
			validKey := true
			if tt.hasUser {
				u = content.User{Login: "test", HashType: "scrypt"}
				if tt.key == passwordSetting {
					if err := u.Password("pass", secret); err != nil {
						t.Fatal(err)
					}
				}
				r = r.WithContext(context.WithValue(r.Context(), userKey, u))
				params := []string{}

				if tt.otherUser {
					params = append(params, "name", "test1")

					var err error

					if tt.otherUserErr {
						err = errors.New("test1")
						code = http.StatusInternalServerError
					} else {
						u = content.User{Login: "test1"}
					}

					userRepo.EXPECT().Get(content.Login("test1")).Return(u, err)
				}

				params = append(params, "key", tt.key)
				r = addChiParam(r, params...)
				switch tt.key {
				case firstNameSetting:
					u.FirstName = r.Form.Get("value")
				case lastNameSetting:
					u.LastName = r.Form.Get("value")
				case emailSetting:
					if !tt.formErr {
						u.Email = r.Form.Get("value")
					}
				case profileSetting:
					if !tt.formErr {
						if err := json.Unmarshal([]byte(r.Form.Get("value")), &u.ProfileData); err != nil {
							t.Fatal(err)
						}
					}
				case activeSetting:
					u.Active = true
				case passwordSetting:
				default:
					validKey = false
					if tt.hasUser && !tt.otherUserErr {
						code = http.StatusNotFound
					}
				}

				if !tt.otherUserErr && validKey && !tt.formErr {
					var err error
					if tt.updateErr {
						err = errors.New("update err")
						code = http.StatusInternalServerError
					} else {
						code = http.StatusOK
					}
					userRepo.EXPECT().Update(userMatcher{u}).DoAndReturn(func(u content.User) error {
						got = u
						return err
					})
				}
			}

			setSettingValue(userRepo, secret, logger).ServeHTTP(w, r)

			if w.Code != code {
				t.Errorf("setSettingValue() code = %v, want %v", w.Code, code)
				return
			}

			exp := content.User{}
			if tt.hasUser && !tt.otherUserErr && validKey && !tt.formErr {
				exp = u
			}

			if tt.key == passwordSetting && !tt.formErr {
				if len(got.Hash) == 0 {
					t.Errorf("setSettingValue() got hash = %v", got.Hash)
					return
				}
			}
			if !(userMatcher{got}).Matches(exp) {
				t.Errorf("setSettingValue() got = %v, want %v", got, exp)
			}
		})
	}
}
