package readeef

import (
	"net/http"

	"github.com/urandom/readeef/content/data"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/renderer"
)

type Login struct {
	webfw.BasePatternController
}

func NewLogin(pattern string) Login {
	return Login{webfw.NewBasePatternController(pattern, webfw.MethodGet|webfw.MethodPost, "auth-login")}
}

func (con Login) Handler(c context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := webfw.GetLogger(c)
		sess := webfw.GetSession(c, r)
		renderData := renderer.RenderData{}

		if r.Method == "GET" {
			if v, ok := sess.Flash("form-error"); ok {
				renderData["form-error"] = v
			}
		} else {
			if err := r.ParseForm(); err != nil {
				l.Fatal(err)
			}
			username := data.Login(r.Form.Get("username"))
			password := r.Form.Get("password")

			repo := GetRepo(c)
			conf := GetConfig(c)
			u := repo.UserByLogin(username)

			formError := false
			if u.Err() != nil {
				sess.SetFlash("form-error", "login-incorrect")
				formError = true
			} else if !u.Authenticate(password, []byte(conf.Auth.Secret)) {
				sess.SetFlash("form-error", "login-incorrect")
				formError = true
			} else {
				sess.Set(AuthUserKey, u)
				sess.Set(AuthNameKey, username)
			}

			if formError {
				http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
			} else {
				var returnPath string
				if v, ok := sess.Flash("return-to"); ok {
					returnPath = v.(string)
				} else {
					returnPath = webfw.GetDispatcher(c).Pattern
				}
				http.Redirect(w, r, returnPath, http.StatusTemporaryRedirect)
			}
			return
		}
		err := webfw.GetRenderCtx(c, r)(w, renderData, "login.tmpl")
		if err != nil {
			l.Print(err)
		}
	}
}
