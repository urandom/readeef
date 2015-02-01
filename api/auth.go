package api

import (
	"encoding/json"
	"net/http"

	"github.com/urandom/readeef"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
)

type Auth struct {
	webfw.BasePatternController
}

func NewAuth() Auth {
	return Auth{
		webfw.NewBasePatternController("/v:version/auth", webfw.MethodAll, ""),
	}
}

func (con Auth) Handler(c context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := readeef.GetUser(c, r)

		type User struct {
			Login     string
			FirstName string
			LastName  string
			Email     string
			Admin     bool
		}
		type response struct {
			Auth        bool
			User        User
			ProfileData map[string]interface{}
		}

		resp := response{
			Auth: true,
			User: User{
				Login:     user.Login,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Email:     user.Email,
				Admin:     user.Admin,
			},
			ProfileData: user.ProfileData,
		}

		b, err := json.Marshal(resp)
		if err != nil {
			webfw.GetLogger(c).Print(err)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(b)
	})
}

func (con Auth) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}
