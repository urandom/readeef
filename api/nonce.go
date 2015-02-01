package api

import (
	"encoding/json"
	"net/http"

	"github.com/urandom/readeef"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
)

type Nonce struct {
	webfw.BasePatternController
	nonce *readeef.Nonce
}

func NewNonce(nonce *readeef.Nonce) Nonce {
	return Nonce{
		webfw.NewBasePatternController("/v:version/nonce", webfw.MethodAll, ""),
		nonce,
	}
}

func (con Nonce) Handler(c context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce := con.nonce.Generate()
		type response struct {
			Nonce string
		}
		resp := response{nonce}

		b, err := json.Marshal(resp)
		if err != nil {
			webfw.GetLogger(c).Print(err)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(b)

		con.nonce.Set(nonce)
	})
}

func (con Nonce) AuthRequired(c context.Context, r *http.Request) bool {
	return false
}
