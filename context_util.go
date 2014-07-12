package readeef

import "github.com/urandom/webfw/context"

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

	db, err := NewDB(conf.DB.Driver, conf.DB.Connect)
	if err != nil {
		panic(err)
	}

	return db
}
