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

// The Auth middleware checks whether the session contains a valid user or
// login. If it only contains the later, it tries to load the actual user
// object from the database. If a valid user hasn't been loaded, it redirects
// to the named controller "auth-login".
//
// The middleware expects a readeef.DB object, as well the dispatcher's
// pattern. It may also receive a slice of path prefixes, relative to the
// dispatcher's pattern, which should be ignored. The later may also be passed
// from the Config.Auth.IgnoreURLPrefix
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
			d := webfw.GetDispatcher(c)
			sess.SetFlash(CtxKey("return-to"), r.URL.Path)
			path := d.NameToPath("auth-login", webfw.MethodGet)

			if path == "" {
				path = "/"
			}

			http.Redirect(w, r, path, http.StatusMovedPermanently)
			return
		}

		ph.ServeHTTP(w, r)
	}

	return http.HandlerFunc(handler)
}
