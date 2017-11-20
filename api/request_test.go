package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/urandom/handler/auth"
	"github.com/urandom/readeef/config"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/mock_repo"
	"github.com/urandom/readeef/log"
)

var (
	cfg    config.Log
	logger log.Log

	secret = []byte("test")
)

func Test_userContext(t *testing.T) {
	tests := []struct {
		name       string
		token      bool
		notFound   bool
		genericErr bool
	}{
		{"no token", false, false, false},
		{"with user", true, false, false},
		{"not found", true, true, false},
		{"generic error", true, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userRepo := mock_repo.NewMockUser(ctrl)

			var url string
			var code int
			if tt.token {
				url = "/?token=" + generateToken("test", time.Now().Add(time.Minute))

				var user content.User
				var err error

				if tt.notFound {
					code = http.StatusNotFound
					err = content.ErrNoContent
				} else if tt.genericErr {
					code = http.StatusInternalServerError
					err = errors.New("test")
				} else {
					user = content.User{Login: "test"}
					code = http.StatusNoContent
				}

				userRepo.EXPECT().Get(content.Login("test")).Return(user, err)
			} else {
				url = "/?token=" + invalidToken()
				code = http.StatusBadRequest
			}
			r := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			auth.RequireToken(
				userContext(userRepo, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				}), logger),
				auth.TokenValidatorFunc(func(token string, claims jwt.Claims) bool {
					return true
				}),
				[]byte("test"),
				auth.Claimer(func(c *jwt.StandardClaims) jwt.Claims {
					if tt.token {
						return c
					}
					return jwt.MapClaims{}
				}),
			).ServeHTTP(w, r)

			if w.Code != code {
				t.Errorf("userContext() = %v, want %v", w.Code, code)
			}
		})
	}
}

func Test_userValidator(t *testing.T) {
	tests := []struct {
		name     string
		wantStop bool
	}{
		{"success", false},
		{"stop", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			if !tt.wantStop {
				r = r.WithContext(context.WithValue(r.Context(), userKey, content.User{Login: "test"}))
			}

			userValidator(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})).ServeHTTP(w, r)

			if tt.wantStop {
				if w.Code != http.StatusBadRequest {
					t.Errorf("userValidator() = %v, want %v", w.Code, http.StatusBadRequest)
				}
			} else {
				if w.Code != http.StatusNoContent {
					t.Errorf("userValidator() = %v, want %v", w.Code, http.StatusNoContent)
				}
			}
		})
	}
}

func Test_userFromRequest(t *testing.T) {
	tests := []struct {
		name     string
		wantUser content.User
		wantStop bool
	}{
		{"with user", content.User{Login: "login"}, false},
		{"without user", content.User{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			if !tt.wantStop {
				r = r.WithContext(context.WithValue(r.Context(), userKey, tt.wantUser))
			}

			gotUser, gotStop := userFromRequest(w, r)
			if !reflect.DeepEqual(gotUser, tt.wantUser) {
				t.Errorf("userFromRequest() gotUser = %v, want %v", gotUser, tt.wantUser)
			}
			if gotStop != tt.wantStop {
				t.Errorf("userFromRequest() gotStop = %v, want %v", gotStop, tt.wantStop)
			}
			if gotStop && w.Code != http.StatusBadRequest {
				t.Errorf("userFromRequest() code = %v, want %v", w.Code, http.StatusBadRequest)
			}
		})
	}
}

func init() {
	cfg.Converted.Writer = os.Stderr
	cfg.Converted.Prefix = "[testing] "

	logger = log.WithStd(cfg)
}

func generateToken(user string, exp time.Time) string {
	t := jwt.NewWithClaims(jwt.SigningMethodHS512, &jwt.StandardClaims{
		Subject:   user,
		ExpiresAt: exp.Unix(),
	})

	if token, err := t.SignedString(secret); err == nil {
		return token
	} else {
		panic(err)
	}
}

func invalidToken() string {
	t := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{})

	if token, err := t.SignedString(secret); err == nil {
		return token
	} else {
		panic(err)
	}
}
