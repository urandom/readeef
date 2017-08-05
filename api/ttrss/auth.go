package ttrss

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
)

func registerAuthActions(sessionManager sessionManager, secret []byte) {
	actions["login"] = func(req request, u content.User, service repo.Service) (interface{}, error) {
		return login(req, u, sessionManager, secret)
	}
	actions["logout"] = func(req request, u content.User, service repo.Service) (interface{}, error) {
		return logout(req, u, sessionManager)
	}
	actions["isLoggedIn"] = func(req request, u content.User, service repo.Service) (interface{}, error) {
		return isLoggedIn(req, u, sessionManager)
	}
}

func login(
	req request, user content.User, sessionManager sessionManager, secret []byte,
) (interface{}, error) {
	if ok, err := user.Authenticate(req.Password, []byte(secret)); !ok {
		return nil, errors.WithStack(newErr(fmt.Sprintf(
			"authentication for TT-RSS user '%s'", user.Login,
		), "LOGIN_ERROR"))
	} else if err != nil {
		return nil, errors.WithStack(newErr(fmt.Sprintf(
			"authentication for TT-RSS user '%s': %v", user.Login, err,
		), "LOGIN_ERROR"))
	}

	sessId := sessionManager.update(session{login: user.Login, lastVisit: time.Now()})

	return genericContent{
		ApiLevel:  API_LEVEL,
		SessionId: sessId,
	}, nil
}

func logout(
	req request, user content.User, sessionManager sessionManager,
) (interface{}, error) {
	sessionManager.remove(req.Sid)
	return genericContent{Status: "OK"}, nil
}

func isLoggedIn(
	req request, user content.User, sessionManager sessionManager,
) (interface{}, error) {
	s := sessionManager.get(req.Sid)
	return genericContent{Status: s.login != ""}, nil
}
