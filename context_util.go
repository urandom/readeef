package readeef

import (
	"net/http"

	"github.com/urandom/webfw/context"
)

type CtxKey string

func GetConfig(c context.Context) Config {
	if v, ok := c.GetGlobal(CtxKey("config")); ok {
		return v.(Config)
	}

	return Config{}
}

func GetDB(c context.Context) DB {
	if v, ok := c.GetGlobal(CtxKey("db")); ok {
		return v.(DB)
	}

	conf := GetConfig(c)

	db := NewDB(conf.DB.Driver, conf.DB.Connect)

	if err := db.Connect(); err != nil {
		panic(err)
	}

	return db
}

func GetUser(c context.Context, r *http.Request) User {
	if v, ok := c.GetGlobal(CtxKey("user")); ok {
		return v.(User)
	}

	return User{}
}
