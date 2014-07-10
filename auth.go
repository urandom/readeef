package readeef

import (
	"log"
	"net/http"
	"strings"

	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
)

const authkey = "AUTHUSER"
const namekey = "AUTHNAME"

type Auth struct {
	DB              DB
	Pattern         string
	IgnoreURLPrefix []string
}

func (mw Auth) Handler(ph http.Handler, c context.Context, l *log.Logger) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		for _, prefix := range mw.IgnoreURLPrefix {
			if prefix[0] == '/' {
				prefix = prefix[1:]
			}

			if strings.HasPrefix(r.URL.Path, mw.Pattern+prefix+"/") {
				ph.ServeHTTP(w, r)
				return
			}
		}
		sess := webfw.GetSession(c, r)

		var u User
		validUser := false
		if uv, ok := sess.Get(authkey); ok {
			if u, ok = uv.(User); ok {
				validUser = true
			}
		}

		if !validUser {
			if uv, ok := sess.Get(namekey); ok {
				if n, ok := uv.(string); ok {
					var err error
					u, err = mw.DB.GetUser(n)

					if err == nil {
						validUser = true
						sess.Set(authkey, u)
					} else if _, ok := err.(ValidationError); !ok {
						l.Print(err)
					}
				}
			}
		}

		if !validUser {
			http.Redirect(w, r, r.URL.String(), http.StatusForbidden)
			return
		}

		ph.ServeHTTP(w, r)
	}

	return http.HandlerFunc(handler)
}
