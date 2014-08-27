package api

import (
	"encoding/json"
	"net/http"

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
		type response struct {
			Auth bool
		}
		resp := response{true}

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
