package readeef

import (
	"net/http"

	"github.com/urandom/readeef/content"
	"github.com/urandom/webfw/context"
)

type CtxKey string

func GetConfig(c context.Context) Config {
	if v, ok := c.GetGlobal(CtxKey("config")); ok {
		return v.(Config)
	}

	return Config{}
}

func GetRepo(c context.Context) content.Repo {
	if v, ok := c.GetGlobal(CtxKey("repo")); ok {
		return v.(content.Repo)
	}

	return nil
}

func GetUser(c context.Context, r *http.Request) content.User {
	if v, ok := c.Get(r, context.BaseCtxKey("user")); ok {
		return v.(content.User)
	}

	return nil
}
