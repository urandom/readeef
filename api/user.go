package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/urandom/readeef"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/util"
)

type User struct{}

type listUsersProcessor struct {
	Attribute string `json:"attribute"`

	db   readeef.DB
	user readeef.User
}

type addUserProcessor struct {
	Login    string `json:"login"`
	Password string `json:"password"`

	db   readeef.DB
	user readeef.User
}

type removeUserProcessor struct {
	Login string `json:"login"`

	db   readeef.DB
	user readeef.User
}

type setAttributeForUserProcessor struct {
	Login     string          `json:"login"`
	Attribute string          `json:"attribute"`
	Value     json.RawMessage `json:"value"`

	db   readeef.DB
	user readeef.User
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db := readeef.GetDB(c)
		user := readeef.GetUser(c, r)

		action := webfw.GetMultiPatternIdentifier(c, r)
		params := webfw.GetParams(c, r)

		var resp responseError
		switch action {
		case "list":
			resp = listUsers(db, user)
		case "add":
			buf := util.BufferPool.GetBuffer()
			defer util.BufferPool.Put(buf)

			buf.ReadFrom(r.Body)

			resp = addUser(db, user, params["login"], buf.String())
		case "remove":
			resp = removeUser(db, user, params["login"])
		case "setAttr":
			resp = setAttributeForUser(db, user, params["login"], params["attr"], []byte(params["value"]))
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
	return listUsers(p.db, p.user)
}

func (p addUserProcessor) Process() responseError {
	return addUser(p.db, p.user, p.Login, p.Password)
}

func (p removeUserProcessor) Process() responseError {
	return removeUser(p.db, p.user, p.Login)
}

func (p setAttributeForUserProcessor) Process() responseError {
	return setAttributeForUser(p.db, p.user, p.Login, p.Attribute, p.Value)
}

func listUsers(db readeef.DB, user readeef.User) (resp responseError) {
	resp = newResponse()

	if !user.Admin {
		resp.err = errForbidden
		resp.errType = errTypeForbidden
		return
	}

	var users []readeef.User
	if users, resp.err = db.GetUsers(); resp.err != nil {
		return
	}

	type respUser struct {
		Login     string
		FirstName string
		LastName  string
		Email     string
		Active    bool
		Admin     bool
	}

	userList := []respUser{}
	for _, u := range users {
		userList = append(userList, respUser{
			Login:     u.Login,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Email:     u.Email,
			Active:    u.Active,
			Admin:     u.Admin,
		})
	}

	resp.val["Users"] = userList
	return
}

func addUser(db readeef.DB, user readeef.User, login, password string) (resp responseError) {
	resp = newResponse()
	resp.val["Login"] = login

	if !user.Admin {
		resp.err = errForbidden
		resp.errType = errTypeForbidden
		return
	}

	_, resp.err = db.GetUser(login)
	if resp.err == nil {
		/* TODO: non-fatal error */
		resp.err = errUserExists
		resp.errType = errTypeUserExists
		return
	} else if resp.err != sql.ErrNoRows {
		return
	}

	resp.err = nil

	u := readeef.User{Login: login}

	if resp.err = u.SetPassword(password); resp.err != nil {
		return
	}

	if resp.err = db.UpdateUser(u); resp.err != nil {
		return
	}

	resp.val["Success"] = true

	return
}

func removeUser(db readeef.DB, user readeef.User, login string) (resp responseError) {
	resp = newResponse()
	resp.val["Login"] = login

	if !user.Admin {
		resp.err = errForbidden
		resp.errType = errTypeForbidden
		return
	}

	if user.Login == login {
		resp.err = errCurrentUser
		resp.errType = errTypeCurrentUser
		return
	}

	var u readeef.User

	if u, resp.err = db.GetUser(login); resp.err != nil {
		return
	}

	if resp.err = db.DeleteUser(u); resp.err != nil {
		return
	}

	resp.val["Success"] = true
	return
}

func setAttributeForUser(db readeef.DB, user readeef.User, login, attr string, value []byte) (resp responseError) {
	if !user.Admin {
		resp.err = errForbidden
		resp.errType = errTypeForbidden
		return
	}

	if user.Login == login {
		resp.err = errCurrentUser
		resp.errType = errTypeCurrentUser
		return
	}

	var u readeef.User

	if u, resp.err = db.GetUser(login); resp.err != nil {
		return
	}

	resp = setUserAttribute(db, u, attr, value)
	return
}
