package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	time "time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/mock_repo"
)

func Test_tokenCreate(t *testing.T) {
	withPass := content.User{Login: "test"}
	withPass.Password("pass", secret)
	tests := []struct {
		name       string
		form       url.Values
		noAuthData bool
		user       content.User
		userErr    error
	}{
		{name: "no auth data", noAuthData: true},
		{name: "user err", form: url.Values{"user": []string{"test"}, "password": []string{"test"}}, userErr: errors.New("err")},
		{name: "invalid password", form: url.Values{"user": []string{"test"}, "password": []string{"wrong"}}, user: withPass},
		{name: "valid password", form: url.Values{"user": []string{"test"}, "password": []string{"pass"}}, user: withPass},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userRepo := mock_repo.NewMockUser(ctrl)

			r := httptest.NewRequest("POST", "/", strings.NewReader(tt.form.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			r.ParseForm()
			w := NewCloseNotifier()

			code := http.StatusUnauthorized
			switch {
			default:
				if tt.noAuthData {
					break
				}

				userRepo.EXPECT().Get(content.Login(tt.form.Get("user"))).Return(tt.user, tt.userErr)

				if tt.userErr != nil {
					break
				}

				if tt.form.Get("password") != "pass" {
					break
				}

				code = http.StatusOK
			}

			tokenCreate(userRepo, secret, logger).ServeHTTP(w, r)

			if code != w.Code {
				t.Errorf("tokenCreate() code = %v, want %v", code, w.Code)
				return
			}

			if strings.HasPrefix(w.Header().Get("Authorization"), "Bearer ") != (w.Code == http.StatusOK) {
				t.Errorf("tokenCreate() authorization header = %v, code %v", w.Header().Get("Authorization"), w.Code)
				return
			}
		})
	}
}

func Test_tokenDelete(t *testing.T) {
	tests := []struct {
		name     string
		tokenFor string
		tokenExp time.Time
		storeErr error
	}{
		{name: "no token"},
		{name: "expired token", tokenFor: "user", tokenExp: time.Now().Add(-100 * time.Minute)},
		{name: "store err", tokenFor: "user", tokenExp: time.Now().Add(100 * time.Minute), storeErr: errors.New("err")},
		{name: "successful delete", tokenFor: "user", tokenExp: time.Now().Add(100 * time.Minute)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			storage := NewMockStorage(ctrl)

			r := httptest.NewRequest("DELETE", "/", nil)
			w := NewCloseNotifier()

			code := http.StatusOK
			now := time.Now()
			switch {
			default:
				token := generateToken(tt.tokenFor, tt.tokenExp)
				if tt.tokenFor != "" {
					r.Header.Set("Authorization", "Bearer "+token)
				} else {
					code = http.StatusBadRequest
					break
				}

				if tt.tokenExp.Sub(now) > 0 {
					storage.EXPECT().Store(token, gomock.Any()).Return(tt.storeErr)
				}

				if tt.storeErr != nil {
					code = http.StatusInternalServerError
					break
				}
			}

			tokenDelete(storage, secret, logger).ServeHTTP(w, r)

			if code != w.Code {
				t.Errorf("tokenDelete() code = %v, want %v", code, w.Code)
				return
			}
		})
	}
}

func Test_tokenValidator(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		tokenExists    bool
		tokenExistsErr error
		claims         jwt.Claims
		subject        string
		userErr        error
		want           bool
	}{
		{name: "no token"},
		{name: "error reading storage", token: "token", tokenExistsErr: errors.New("err")},
		{name: "does not exist", token: "token", tokenExists: true},
		{name: "not standard claims", token: "token", claims: claims{}},
		{name: "user does not exist", token: "token", claims: &jwt.StandardClaims{Subject: "user"}, subject: "user", userErr: content.ErrNoContent},
		{name: "error getting user", token: "token", claims: &jwt.StandardClaims{Subject: "user"}, subject: "user", userErr: errors.New("err")},
		{name: "valid user", token: "token", claims: &jwt.StandardClaims{Subject: "user"}, subject: "user", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			storage := NewMockStorage(ctrl)
			userRepo := mock_repo.NewMockUser(ctrl)

			storage.EXPECT().Exists(tt.token).Return(tt.tokenExists, tt.tokenExistsErr)

			if tt.subject != "" {
				userRepo.EXPECT().Get(content.Login(tt.subject)).Return(content.User{}, tt.userErr)
			}

			got := tokenValidator(userRepo, storage, logger).Validate(tt.token, tt.claims)

			if got != tt.want {
				t.Errorf("tokenValidator() = %v, want %v", got, tt.want)
				return
			}
		})
	}
}
