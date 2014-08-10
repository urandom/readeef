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

func NewAuth() Feed {
	return Auth{
		webfw.NewBaseController("/v:version/auth", webfw.MethodAll, ""),
	}
}

func (con Feed) Handler(c context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type response struct {
			auth bool
		}
		resp := response{true}

		var b []byte
		b, err = json.Marshal(resp)
		if err != nil {
			webfw.GetLogger(c).Print(err)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(b)
	}
}

func (con Feed) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}
