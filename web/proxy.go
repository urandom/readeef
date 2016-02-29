package web

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/urandom/readeef"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
)

type Proxy struct {
	webfw.BasePatternController
}

func NewProxy() Proxy {
	return Proxy{
		BasePatternController: webfw.NewBasePatternController("/proxy", webfw.MethodGet, ""),
	}
}

func (con Proxy) Handler(c context.Context) http.Handler {
	logger := webfw.GetLogger(c)
	config := readeef.GetConfig(c)
	client := readeef.NewTimeoutClient(config.Timeout.Converted.Connect, config.Timeout.Converted.ReadWrite)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess := webfw.GetSession(c, r)

		if _, ok := sess.Get(readeef.AuthNameKey); !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if r.Method == "HEAD" {
			return
		}

		r.ParseForm()
		var err error

		switch {
		default:
			var u *url.URL

			u, err = url.Parse(r.Form.Get("url"))
			if err != nil {
				err = fmt.Errorf("Error parsing url to proxy (%s): %v", r.Form.Get("url"), err)
				break
			}
			if u.Scheme == "" {
				u.Scheme = "http"
			}

			var req *http.Request

			req, err = http.NewRequest("GET", u.String(), nil)
			if err != nil {
				err = fmt.Errorf("Error creating proxy request to %s: %v", u, err)
				break
			}

			var resp *http.Response

			resp, err = client.Do(req)
			if err != nil {
				err = fmt.Errorf("Error getting proxy response from %s: %v", u, err)
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
				err = fmt.Errorf("Error reading proxy response from %s: %v", u, err)
				break
			}

			_, err = w.Write(b)
		}

		if err != nil {
			logger.Infoln(err)
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}

		return
	})
}
