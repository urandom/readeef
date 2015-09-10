package readeef

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/util"
)

const AuthUserKey = "AUTHUSER"
const AuthNameKey = "AUTHNAME"

type AuthController interface {
	LoginRequired(context.Context, *http.Request) bool
}

type ApiAuthController interface {
	AuthRequired(context.Context, *http.Request) bool
}

type AuthRejectHandler interface {
	AuthReject(context.Context, *http.Request)
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
	Nonce           *Nonce
	Pattern         string
	IgnoreURLPrefix []string
}

var (
	loginRegexp     = regexp.MustCompile(`\blogin=.*?(?:&|$)`)
	signatureRegexp = regexp.MustCompile(`\bsignature=.*?(?:&|$)`)
	dateRegexp      = regexp.MustCompile(`\bdate=.*?(?:&|$)`)
	nonceRegexp     = regexp.MustCompile(`\bnonce=.*?(?:&|$)`)
)

func (mw Auth) Handler(ph http.Handler, c context.Context) http.Handler {
	logger := webfw.GetLogger(c)
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

		repo := GetRepo(c)

		switch ac := route.Controller.(type) {
		case AuthController:
			if !ac.LoginRequired(c, r) {
				ph.ServeHTTP(w, r)
				return
			}

			sess := webfw.GetSession(c, r)

			var u content.User
			validUser := false
			if uv, ok := sess.Get(AuthUserKey); ok {
				if u, ok = uv.(content.User); ok {
					validUser = true
				}
			}

			if !validUser {
				if uv, ok := sess.Get(AuthNameKey); ok {
					if n, ok := uv.(data.Login); ok {
						u = repo.UserByLogin(n)

						if u.HasErr() {
							logger.Print(u.Err())
						} else {
							validUser = true
							sess.Set(AuthUserKey, u)
						}
					}
				}
			}

			if validUser && !u.Data().Active {
				logger.Infoln("User " + u.Data().Login + " is inactive")
				validUser = false
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

			url, login, signature, nonce, date, t := authData(r)

			validUser := false

			var u content.User

			if login != "" && signature != "" && !t.IsZero() {
				switch {
				default:
					u = repo.UserByLogin(data.Login(login))
					if u.HasErr() {
						logger.Printf("Error getting db user '%s': %v\n", login, u.Err())
						break
					}

					decoded, err := base64.StdEncoding.DecodeString(signature)
					if err != nil {
						logger.Printf("Error decoding auth header: %v\n", err)
						break
					}

					if t.Add(30 * time.Second).Before(time.Now()) {
						break
					}

					if !mw.Nonce.Check(nonce) {
						break
					}
					mw.Nonce.Remove(nonce)

					buf := util.BufferPool.GetBuffer()
					defer util.BufferPool.Put(buf)

					buf.ReadFrom(r.Body)
					r.Body = ioutil.NopCloser(buf)

					bodyHash := md5.New()
					if _, err := bodyHash.Write(buf.Bytes()); err != nil {
						logger.Printf("Error generating the hash for the request body: %v\n", err)
						break
					}

					contentMD5 := base64.StdEncoding.EncodeToString(bodyHash.Sum(nil))

					message := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n",
						url, r.Method, contentMD5, r.Header.Get("Content-Type"),
						date, nonce)

					b := make([]byte, base64.StdEncoding.EncodedLen(len(u.Data().MD5API)))
					base64.StdEncoding.Encode(b, u.Data().MD5API)

					hm := hmac.New(sha256.New, b)
					if _, err := hm.Write([]byte(message)); err != nil {
						logger.Printf("Error generating the hashed message: %v\n", err)
						break
					}

					if !hmac.Equal(hm.Sum(nil), decoded) {
						logger.Printf("Error matching the supplied auth message to the generated one.\n")
						break
					}

					if !u.Data().Active {
						logger.Println("User " + u.Data().Login + " is inactive")
						break
					}

					validUser = true
				}
			}

			if validUser {
				c.Set(r, context.BaseCtxKey("user"), u)
			} else {
				if rej, ok := ac.(AuthRejectHandler); ok {
					rej.AuthReject(c, r)
				} else {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
			}
		}

		ph.ServeHTTP(w, r)
	}

	return http.HandlerFunc(handler)
}

func authData(r *http.Request) (string, string, string, string, string, time.Time) {
	var login, signature string
	var t time.Time

	auth := r.Header.Get("Authorization")
	url := r.RequestURI

	if auth == "" {
		login = r.FormValue("login")
		signature = r.FormValue("signature")

		url = loginRegexp.ReplaceAllString(url, "")
		url = signatureRegexp.ReplaceAllString(url, "")
	} else {
		if strings.HasPrefix(auth, "Readeef ") {
			auth = auth[len("Readeef "):]

			parts := strings.SplitN(auth, ":", 2)
			login, signature = parts[0], parts[1]
		}
	}

	nonce := r.Header.Get("X-Nonce")
	if nonce == "" {
		nonce = r.FormValue("nonce")

		url = nonceRegexp.ReplaceAllString(url, "")
	}

	date := r.Header.Get("Date")

	if date == "" {
		date = r.Header.Get("X-Date")
	}

	if date == "" {
		date = r.FormValue("date")

		url = dateRegexp.ReplaceAllString(url, "")
	}

	dateInt, err := strconv.ParseInt(date, 10, 64)
	if err == nil {
		t = time.Unix(dateInt, 0)
	} else {
		t, _ = time.Parse(http.TimeFormat, date)
	}

	if url != r.RequestURI {
		if strings.HasSuffix(url, "?") || strings.HasSuffix(url, "&") {
			url = url[:len(url)-1]
		}
	}

	return url, login, signature, nonce, date, t
}
