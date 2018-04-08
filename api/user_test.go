package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/mock_repo"
)

func Test_getUserData(t *testing.T) {
	tests := []struct {
		name string
		user bool
	}{
		{"has user", false},
		{"no user", true},
	}
	type data struct {
		User content.User `json:"user"`
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			code := http.StatusBadRequest
			var user content.User
			if tt.user {
				code = http.StatusOK
				user = content.User{Login: "test"}
				r = r.WithContext(context.WithValue(r.Context(), userKey, user))
			}
			getUserData(w, r)

			if w.Code != code {
				t.Errorf("getUserData() code = %v, want %v", w.Code, code)
				return
			}

			if tt.user {
				got := data{}
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Errorf("getUserData() body = '%s', error = %v", w.Body, err)
					return
				}

				if !reflect.DeepEqual(got.User, user) {
					t.Errorf("getUserData() user = %v, want = %v", got.User, user)
				}
			}
		})
	}
}

func Test_createUserToken(t *testing.T) {
	tests := []struct {
		name   string
		noUser bool
		code   int
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			switch {
			default:
				if tt.noUser {
					break
				}

				r = r.WithContext(context.WithValue(r.Context(), userKey, content.User{Login: "test"}))
			}

			createUserToken(secret, logger).ServeHTTP(w, r)

			if tt.code != w.Code {
				t.Errorf("createUserToken() code = %v, want %v", w.Code, tt.code)
				return
			}

			if strings.HasPrefix(w.Header().Get("Authorization"), "Bearer ") != (w.Code == http.StatusOK) {
				t.Errorf("createUserToken() authorization header = %v, code %v", w.Header().Get("Authorization"), w.Code)
				return
			}
		})
	}
}

func Test_listUsers(t *testing.T) {
	tests := []struct {
		name  string
		user  bool
		users int
		error bool
	}{
		{"no user", false, 0, false},
		{"single user", true, 1, false},
		{"three user", true, 3, false},
		{"repo error", true, 0, true},
	}
	type data struct {
		Users []content.User `json:"users"`
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userRepo := mock_repo.NewMockUser(ctrl)

			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			code := http.StatusBadRequest
			var users []content.User
			if tt.user {
				r = r.WithContext(context.WithValue(r.Context(), userKey, content.User{Login: "test"}))

				var err error
				for i := 0; i < tt.users; i++ {
					users = append(users, content.User{Login: content.Login(fmt.Sprintf("test%d", i)), FirstName: "tester"})
				}
				if tt.error {
					err = errors.New("test")
					code = http.StatusInternalServerError
				} else {
					code = http.StatusOK
				}

				userRepo.EXPECT().All().Return(users, err)
			}

			listUsers(userRepo, logger).ServeHTTP(w, r)

			if w.Code != code {
				t.Errorf("listUsers() code = %v, want %v", w.Code, code)
				return
			}

			if tt.user && !tt.error {
				got := data{}
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Errorf("getUserData() body = '%s', error = %v", w.Body, err)
					return
				}

				if !reflect.DeepEqual(got.Users, users) {
					t.Errorf("getUserData() users = %v, want = %v", got.Users, users)
				}
			}
		})
	}
}

func Test_addUser(t *testing.T) {
	withPass := content.User{Login: "test1", Email: "test@example.com", Active: true}
	withPass.Password("pass", secret)
	full := content.User{Login: "test1", FirstName: "first", LastName: "last", Email: "test@example.com", Active: true, Admin: true}
	full.Password("pass", secret)

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	w.WriteField("login", "test1")
	w.WriteField("email", "test@example.com")
	w.WriteField("active", "")
	w.WriteField("password", "pass")
	w.Close()

	urlEnc := "application/x-www-form-urlencoded"
	tests := []struct {
		name      string
		login     string
		form      io.Reader
		cType     string
		user      bool
		exists    bool
		existsErr bool
		error     bool
		added     content.User
	}{
		{"no user", "test1", strings.NewReader("login=test1"), urlEnc, false, false, false, false, content.User{}},
		{"already exists", "test1", strings.NewReader("login=test1"), urlEnc, true, true, false, false, content.User{}},
		{"exists err", "test1", strings.NewReader("login=test1"), urlEnc, true, false, true, false, content.User{}},
		{"update err", "test1", strings.NewReader("login=test1"), urlEnc, true, false, false, true, content.User{Login: "test1"}},
		{"user with email and pass", "test1", strings.NewReader("login=test1&email=test@example.com&password=pass&active"), urlEnc, true, false, false, false, withPass},
		{"multipart", "test1", &body, w.FormDataContentType(), true, false, false, false, withPass},
		{"full user form", "test1", strings.NewReader("login=test1&firstName=first&lastName=last&email=test@example.com&password=pass&active&admin"), urlEnc, true, false, false, false, full},
	}
	type data struct {
		Success bool `json:"success"`
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userRepo := mock_repo.NewMockUser(ctrl)

			r := httptest.NewRequest("POST", "/", tt.form)
			r.Header.Set("content-type", tt.cType)
			if strings.Contains(tt.cType, "multipart") {
				if err := r.ParseMultipartForm(0); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := r.ParseForm(); err != nil {
					t.Fatal(err)
				}
			}
			w := httptest.NewRecorder()

			code := http.StatusBadRequest
			var got content.User
			if tt.user {
				r = r.WithContext(context.WithValue(r.Context(), userKey, content.User{Login: "test"}))

				var existErr error
				if tt.exists {
					code = http.StatusConflict
				} else if tt.existsErr {
					code = http.StatusInternalServerError
					existErr = errors.New("exists")
				} else {
					existErr = content.ErrNoContent
				}

				userRepo.EXPECT().Get(content.Login(tt.login)).Return(content.User{}, existErr)

				if !tt.exists && !tt.existsErr {
					var err error
					if tt.error {
						err = errors.New("test")
						code = http.StatusInternalServerError
					} else {
						code = http.StatusOK
					}

					userRepo.EXPECT().Update(userMatcher{tt.added}).DoAndReturn(func(u content.User) error {
						got = u
						return err
					})
				}
			}

			addUser(userRepo, secret, logger).ServeHTTP(w, r)

			if w.Code != code {
				t.Errorf("addUser() code = %v, want %v", w.Code, code)
				return
			}

			if !(userMatcher{got}).Matches(tt.added) {
				t.Errorf("addUser() user = %v, want = %v", got, tt.added)
			}

			if tt.user && !tt.exists && !tt.existsErr && !tt.error {
				g := data{}
				if err := json.Unmarshal(w.Body.Bytes(), &g); err != nil {
					t.Errorf("addUser() body = '%s', error = %v", w.Body, err)
					return
				}

				if !g.Success {
					t.Errorf("addUser() success = %v", g.Success)
					return
				}
			}
		})
	}
}

func Test_deleteUser(t *testing.T) {
	tests := []struct {
		name      string
		login     string
		user      bool
		existsErr bool
		error     bool
		deleted   content.User
	}{
		{"no user", "test1", false, false, false, content.User{}},
		{"current user", "test", true, false, false, content.User{}},
		{"exists err", "test1", true, true, false, content.User{}},
		{"delete err", "test1", true, false, true, content.User{Login: "test1"}},
		{"success", "test1", true, false, false, content.User{Login: "test1"}},
	}
	type data struct {
		Success bool `json:"success"`
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userRepo := mock_repo.NewMockUser(ctrl)

			r := httptest.NewRequest("DELETE", "/", nil)
			w := httptest.NewRecorder()

			code := http.StatusBadRequest
			var got content.User
			if tt.user {
				r = r.WithContext(context.WithValue(r.Context(), userKey, content.User{Login: "test"}))
				r = addChiParam(r, "name", tt.login)

				var existErr error
				if tt.login == "test" {
					code = http.StatusConflict
				} else if tt.existsErr {
					code = http.StatusInternalServerError
					existErr = errors.New("exists")
				} else {
					existErr = nil
				}

				if tt.login != "test" {
					userRepo.EXPECT().Get(content.Login(tt.login)).Return(tt.deleted, existErr)
				}

				if tt.login != "test" && !tt.existsErr {
					var err error
					if tt.error {
						err = errors.New("test")
						code = http.StatusInternalServerError
					} else {
						code = http.StatusOK
					}

					userRepo.EXPECT().Delete(userMatcher{tt.deleted}).DoAndReturn(func(u content.User) error {
						got = u
						return err
					})
				}
			}

			deleteUser(userRepo, logger).ServeHTTP(w, r)

			if w.Code != code {
				t.Errorf("deleteUser() code = %v, want %v", w.Code, code)
				return
			}

			if !(userMatcher{got}).Matches(tt.deleted) {
				t.Errorf("deleteUser() user = %v, want = %v", got, tt.deleted)
			}

			if tt.user && tt.login != "test" && !tt.existsErr && !tt.error {
				g := data{}
				if err := json.Unmarshal(w.Body.Bytes(), &g); err != nil {
					t.Errorf("addUser() body = '%s', error = %v", w.Body, err)
					return
				}

				if !g.Success {
					t.Errorf("addUser() success = %v", g.Success)
					return
				}
			}
		})
	}
}

func Test_adminValidator(t *testing.T) {
	tests := []struct {
		name  string
		user  bool
		admin bool
	}{
		{"no user", false, false},
		{"regular user", true, false},
		{"admin user", true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			code := http.StatusBadRequest
			if tt.user {
				u := content.User{Login: "test"}
				code = http.StatusForbidden
				if tt.admin {
					u.Admin = true
					code = http.StatusNoContent
				}
				r = r.WithContext(context.WithValue(r.Context(), userKey, u))
			}

			adminValidator(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})).ServeHTTP(w, r)

			if w.Code != code {
				t.Errorf("adminValidator() code = %d, want %d", w.Code, code)
			}
		})
	}
}

type userMatcher struct{ user content.User }

func (u userMatcher) Matches(x interface{}) bool {
	if user, ok := x.(content.User); ok {
		return u.user.Login == user.Login &&
			u.user.FirstName == user.FirstName &&
			u.user.LastName == user.LastName &&
			u.user.Email == user.Email &&
			u.user.Active == user.Active &&
			u.user.Admin == user.Admin
	}
	return false
}

func (u userMatcher) String() string {
	return "Matches by certain user fields"
}

func addChiParam(r *http.Request, params ...string) *http.Request {
	c := chi.NewRouteContext()
	for i := 0; i < len(params); i += 2 {
		c.URLParams.Add(params[i], params[i+1])
	}

	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, c))
}
