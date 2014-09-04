package api

import (
	"encoding/json"
	"net/http"
	"readeef"

	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
)

type Auth struct {
	webfw.BaseController
}

func NewAuth() Auth {
	return Auth{
		webfw.NewBaseController("/v:version/auth", webfw.MethodAll, ""),
	}
}

func (con Auth) Handler(c context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := readeef.GetUser(c, r)
		if !user.Active {
			readeef.Debug.Println("User " + user.Login + " is inactive")
			w.WriteHeader(http.StatusForbidden)
			return
		}

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
	}
}

func (con Auth) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}
