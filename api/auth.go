package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
)

type Auth struct {
	capabilities capabilities
}

type getAuthDataProcessor struct {
	user         content.User
	session      context.Session
	capabilities capabilities
}

type logoutProcessor struct {
	session context.Session
}

type capabilities struct {
	I18N       bool
	Search     bool
	Extractor  bool
	ProxyHTTP  bool
	Popularity bool
}

func NewAuth(capabilities capabilities) Auth {
	return Auth{
		capabilities,
	}
}

func (con Auth) Patterns() []webfw.MethodIdentifierTuple {
	prefix := "/v:version/auth"

	return []webfw.MethodIdentifierTuple{
		webfw.MethodIdentifierTuple{prefix, webfw.MethodGet, "auth-data"},
		webfw.MethodIdentifierTuple{prefix + "/logout", webfw.MethodPost, "logout"},
	}
}

func (con Auth) Handler(c context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		action := webfw.GetMultiPatternIdentifier(c, r)
		sess := webfw.GetSession(c, r)

		var resp responseError

		switch action {
		case "auth-data":
			user := readeef.GetUser(c, r)
			resp = getAuthData(user, sess, con.capabilities)
		case "logout":
			resp = logout(sess)
		}

		var b []byte
		if resp.err == nil {
			b, resp.err = json.Marshal(resp.val)
		}
		if resp.err == nil {
			w.Write(b)
		} else {
			webfw.GetLogger(c).Print(resp.err)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

func (con Auth) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}

func (p getAuthDataProcessor) Process() responseError {
	return getAuthData(p.user, p.session, p.capabilities)
}

func (p logoutProcessor) Process() responseError {
	return logout(p.session)
}

func getAuthData(user content.User, sess context.Session, capabilities capabilities) (resp responseError) {
	resp = newResponse()

	if sess != nil {
		sess.Set(readeef.AuthNameKey, user.Data().Login)
		if err := sess.Write(nil); err != nil {
			resp.err = fmt.Errorf("Error writing session data: %v", err)
		}
	}

	resp.val["Auth"] = true
	resp.val["Capabilities"] = capabilities
	resp.val["User"] = user
	return
}

func logout(sess context.Session) (resp responseError) {
	resp = newResponse()

	if sess != nil {
		sess.DeleteAll()
		if err := sess.Write(nil); err != nil {
			resp.err = fmt.Errorf("Error writing session data: %v", err)
		}
	}

	return
}
