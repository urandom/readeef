package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/urandom/readeef"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/util"
)

type User struct{}

var (
	errForbidden   = errors.New("Forbidden")
	errUserExists  = errors.New("User exists")
	errCurrentUser = errors.New("Current user")

	errTypeUserExists  = "error-user-exists"
	errTypeCurrentUser = "error-current-user"
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
			resp = addUser(db, user, params["login"], r.Body)
		case "remove":
			resp = removeUser(db, user, params["login"])
		case "setAttr":
			resp = setUserAdminAttribute(db, user, params["login"], params["attr"], params["value"])
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

func listUsers(db readeef.DB, user readeef.User) (resp responseError) {
	resp = newResponse()

	if !user.Admin {
		resp.err = errForbidden
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

func addUser(db readeef.DB, user readeef.User, login string, data io.Reader) (resp responseError) {
	resp = newResponse()
	resp.val["Login"] = login

	if !user.Admin {
		resp.err = errForbidden
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

	buf := util.BufferPool.GetBuffer()
	defer util.BufferPool.Put(buf)

	buf.ReadFrom(data)

	u := readeef.User{Login: login}

	if resp.err = u.SetPassword(buf.String()); resp.err != nil {
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

func setUserAdminAttribute(db readeef.DB, user readeef.User, login, attr, value string) (resp responseError) {
	if !user.Admin {
		resp.err = errForbidden
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

	resp = setUserAttribute(db, u, attr, strings.NewReader(value))
	return
}
