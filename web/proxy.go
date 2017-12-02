package web

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/alexedwards/scs"
	"github.com/pkg/errors"
)

func ProxyHandler(sessionManager *scs.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := sessionManager.Load(r)
		if ok, err := session.GetBool(visitorKey); !ok || err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		r.ParseForm()
		var err error

		switch {
		default:
			var u *url.URL

			u, err = url.Parse(r.Form.Get("url"))
			if err != nil {
				err = errors.Wrapf(err, "parsing url to proxy (%s)", r.Form.Get("url"))
				break
			}
			if u.Scheme == "" {
				u.Scheme = "http"
			}

			var req *http.Request

			req, err = http.NewRequest("GET", u.String(), nil)
			if err != nil {
				err = errors.Wrapf(err, "creating proxy request to %s", u)
				break
			}

			ctx, cancel := context.WithTimeout(req.Context(), 30*time.Second)
			defer cancel()

			var resp *http.Response

			resp, err = http.DefaultClient.Do(req.WithContext(ctx))
			if err != nil {
				err = errors.Wrapf(err, "Error getting proxy response from %s", u)
				break
			}

			defer resp.Body.Close()

			for k, values := range resp.Header {
				for _, v := range values {
					w.Header().Add(k, v)
				}
			}

			var b []byte

			b, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				err = errors.Wrapf(err, "reading proxy response from %s", u)
				break
			}

			_, err = w.Write(b)
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusNotAcceptable)
			return
		}
	}
}
