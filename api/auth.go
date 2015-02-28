package api

import (
	"encoding/json"
	"net/http"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
)

type Auth struct {
	webfw.BasePatternController
}

type getAuthDataProcessor struct {
	user content.User
}

func NewAuth() Auth {
	return Auth{
		webfw.NewBasePatternController("/v:version/auth", webfw.MethodAll, ""),
	}
}

func (con Auth) Handler(c context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := readeef.GetUser(c, r)

		resp := getAuthData(user)

		b, err := json.Marshal(resp.val)
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

func (p getAuthDataProcessor) Process() responseError {
	return getAuthData(p.user)
}

func getAuthData(user content.User) (resp responseError) {
	resp = newResponse()

	in := user.Info()
	resp.val["Auth"] = true
	resp.val["User"] = user
	resp.val["ProfileData"] = in.ProfileData
	return
}
