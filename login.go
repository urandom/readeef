package readeef

import (
	"net/http"

	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/renderer"
)

type Login struct {
	webfw.BaseController
}

func NewLogin(pattern string) Login {
	return Login{webfw.NewBaseController(pattern, webfw.MethodGet|webfw.MethodPost, "auth-login")}
}

func (con Login) Handler(c context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := webfw.GetLogger(c)
		sess := webfw.GetSession(c, r)
		data := renderer.RenderData{}

		if r.Method == "GET" {
			if v, ok := sess.Flash("form-error"); ok {
				data["form-error"] = v
			}
		} else {
			if err := r.ParseForm(); err != nil {
				l.Fatal(err)
			}
			username := r.Form.Get("username")
			password := r.Form.Get("password")

			db := GetDB(c)

			formError := false
			if u, err := db.GetUser(username); err != nil {
				sess.SetFlash("form-error", "login-incorrect")
				formError = true
			} else if !u.Authenticate(password) {
				sess.SetFlash("form-error", "login-incorrect")
				formError = true
			} else {
				sess.Set(authkey, u)
				sess.Set(namekey, u.Login)
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
		err := webfw.GetRenderCtx(c, r)(w, data, "login.tmpl")
		if err != nil {
			l.Print(err)
		}
	}
}
