package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/util"
)

type User struct{}

type listUsersProcessor struct {
	Attribute string `json:"attribute"`

	user content.User
}

type addUserProcessor struct {
	Login    data.Login `json:"login"`
	Password string     `json:"password"`

	user   content.User
	secret []byte
}

type removeUserProcessor struct {
	Login data.Login `json:"login"`

	user content.User
}

type setAttributeForUserProcessor struct {
	Login     data.Login      `json:"login"`
	Attribute string          `json:"attribute"`
	Value     json.RawMessage `json:"value"`

	user   content.User
	secret []byte
}

var (
	errUserExists  = errors.New("User exists")
	errCurrentUser = errors.New("Current user")
	errForbidden   = errors.New("Forbidden")

	errTypeUserExists  = "error-user-exists"
	errTypeCurrentUser = "error-current-user"
	errTypeForbidden   = "error-forbidden"
)

func NewUser() User {
	return User{}
}

func (con User) Patterns() []webfw.MethodIdentifierTuple {
	prefix := "/v:version/user/"

	return []webfw.MethodIdentifierTuple{
		webfw.MethodIdentifierTuple{prefix, webfw.MethodGet, "list"},
		webfw.MethodIdentifierTuple{prefix + ":login", webfw.MethodPost, "add"},
		webfw.MethodIdentifierTuple{prefix + ":login", webfw.MethodDelete, "remove"},
		webfw.MethodIdentifierTuple{prefix + ":login/:attr/:value", webfw.MethodPost, "setAttr"},
	}
}

func (con User) Handler(c context.Context) http.Handler {
	cfg := readeef.GetConfig(c)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := readeef.GetUser(c, r)

		action := webfw.GetMultiPatternIdentifier(c, r)
		params := webfw.GetParams(c, r)

		var resp responseError
		switch action {
		case "list":
			resp = listUsers(user)
		case "add":
			buf := util.BufferPool.GetBuffer()
			defer util.BufferPool.Put(buf)

			buf.ReadFrom(r.Body)

			resp = addUser(user, data.Login(params["login"]), buf.String(), []byte(cfg.Auth.Secret))
		case "remove":
			resp = removeUser(user, data.Login(params["login"]))
		case "setAttr":
			resp = setAttributeForUser(user, []byte(cfg.Auth.Secret), data.Login(params["login"]), params["attr"], []byte(params["value"]))
		}

		switch resp.err {
		case errForbidden:
			w.WriteHeader(http.StatusForbidden)
			return
		case errUserExists:
			resp.val["Error"] = true
			resp.val["ErrorType"] = resp.errType
			resp.err = nil
		case errCurrentUser:
			resp.val["Error"] = true
			resp.val["ErrorType"] = resp.errType
			resp.err = nil
		}

		var b []byte
		if resp.err == nil {
			b, resp.err = json.Marshal(resp.val)
		}
		if resp.err != nil {
			webfw.GetLogger(c).Print(resp.err)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(b)
	})
}

func (con User) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}

func (p listUsersProcessor) Process() responseError {
	return listUsers(p.user)
}

func (p addUserProcessor) Process() responseError {
	return addUser(p.user, p.Login, p.Password, p.secret)
}

func (p removeUserProcessor) Process() responseError {
	return removeUser(p.user, p.Login)
}

func (p setAttributeForUserProcessor) Process() responseError {
	return setAttributeForUser(p.user, p.secret, p.Login, p.Attribute, p.Value)
}

func listUsers(user content.User) (resp responseError) {
	resp = newResponse()

	if !user.Data().Admin {
		resp.err = errForbidden
		resp.errType = errTypeForbidden
		return
	}

	repo := user.Repo()
	resp.val["Users"], resp.err = repo.AllUsers(), repo.Err()
	return
}

func addUser(user content.User, login data.Login, password string, secret []byte) (resp responseError) {
	resp = newResponse()
	resp.val["Login"] = login

	if !user.Data().Admin {
		resp.err = errForbidden
		resp.errType = errTypeForbidden
		return
	}

	repo := user.Repo()
	u := repo.UserByLogin(login)

	if !u.HasErr() {
		/* TODO: non-fatal error */
		resp.err = errUserExists
		resp.errType = errTypeUserExists
		return
	} else {
		err := u.Err()
		if err != content.ErrNoContent {
			resp.err = err
			return
		}
	}

	resp.err = nil

	in := data.User{Login: login}
	u = repo.User()

	u.Data(in)
	u.Password(password, secret)
	u.Update()

	if resp.err = u.Err(); resp.err != nil {
		return
	}

	resp.val["Success"] = true

	return
}

func removeUser(user content.User, login data.Login) (resp responseError) {
	resp = newResponse()
	resp.val["Login"] = login

	if !user.Data().Admin {
		resp.err = errForbidden
		resp.errType = errTypeForbidden
		return
	}

	if user.Data().Login == login {
		resp.err = errCurrentUser
		resp.errType = errTypeCurrentUser
		return
	}

	u := user.Repo().UserByLogin(login)
	u.Delete()

	if resp.err = u.Err(); resp.err != nil {
		return
	}

	resp.val["Success"] = true
	return
}

func setAttributeForUser(user content.User, secret []byte, login data.Login, attr string, value []byte) (resp responseError) {
	if !user.Data().Admin {
		resp.err = errForbidden
		resp.errType = errTypeForbidden
		return
	}

	if user.Data().Login == login {
		resp.err = errCurrentUser
		resp.errType = errTypeCurrentUser
		return
	}

	if u := user.Repo().UserByLogin(login); u.HasErr() {
		resp.err = u.Err()
		return
	} else {
		resp = setUserAttribute(u, secret, attr, value)
	}

	return
}
