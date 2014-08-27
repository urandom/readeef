package readeef

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/util"
)

const authkey = "AUTHUSER"
const namekey = "AUTHNAME"

type AuthController interface {
	LoginRequired(context.Context, *http.Request) bool
}

type ApiAuthController interface {
	AuthRequired(context.Context, *http.Request) bool
}

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
	Nonce           *Nonce
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

		route, _, ok := webfw.GetDispatcher(c).RequestRoute(r)
		if !ok {
			ph.ServeHTTP(w, r)
			return
		}

		switch ac := route.Controller.(type) {
		case AuthController:
			if !ac.LoginRequired(c, r) {
				ph.ServeHTTP(w, r)
				return
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

		case ApiAuthController:
			if !ac.AuthRequired(c, r) {
				ph.ServeHTTP(w, r)
				return
			}

			var u User
			var err error

			auth := r.Header.Get("Authorization")
			validUser := false

			if auth != "" {
				switch {
				default:
					if !strings.HasPrefix(auth, "Readeef ") {
						break
					}
					auth = auth[len("Readeef "):]

					parts := strings.SplitN(auth, ":", 2)
					login := parts[0]

					u, err = mw.DB.GetUser(login)
					if err != nil {
						l.Printf("Error getting db user '%s': %v\n", login, err)
						break
					}

					var decoded []byte
					decoded, err = base64.StdEncoding.DecodeString(parts[1])
					if err != nil {
						l.Printf("Error decoding auth header: %v\n", err)
						break
					}

					date := r.Header.Get("X-Date")
					t, err := time.Parse(http.TimeFormat, date)
					if err != nil || t.Add(30*time.Second).Before(time.Now()) {
						break
					}

					buf := util.BufferPool.GetBuffer()
					defer util.BufferPool.Put(buf)

					buf.ReadFrom(r.Body)
					r.Body = ioutil.NopCloser(buf)

					contentMD5 := base64.StdEncoding.EncodeToString(md5.New().Sum(buf.Bytes()))

					uriParts := strings.SplitN(r.RequestURI, "?", 2)
					uri := uriParts[0]

					message := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n",
						uri, r.Method, contentMD5, r.Header.Get("Content-Type"),
						date)

					b := make([]byte, base64.StdEncoding.EncodedLen(len(u.MD5API)))
					base64.StdEncoding.Encode(b, u.MD5API)

					hm := hmac.New(sha256.New, b)
					if _, err := hm.Write([]byte(message)); err != nil {
						break
					}

					if !hmac.Equal(hm.Sum(nil), decoded) {
						break
					}

					validUser = true
				}
			}

			if !validUser {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			c.Set(r, context.BaseCtxKey("user"), u)
		}

		ph.ServeHTTP(w, r)
	}

	return http.HandlerFunc(handler)
}
