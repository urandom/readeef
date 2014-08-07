package v1

import (
	"errors"
	"net/http"
	"readeef"
	"strings"

	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
)

type Feed struct {
	webfw.BaseController
	db readeef.DB
}

func NewFeed(db readeef.DB) Feed {
	return Feed{
		BaseController: webfw.NewBaseController("/feed/*action", webfw.MethodAll, ""),
		db:             db,
	}
}

func (con Feed) Handler(c context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		actionParam := webfw.GetParams(c, r)
		parts := strings.SplitN(actionParam["action"], "/", 2)
		action, extra := parts[0], parts[1]
		extraParams := strings.Split(extra, "/")

		switch action {
		default:
			err = errors.New("Unknown action " + action)
		}

		if err != nil {
			webfw.GetLogger(c).Print(err)
		}
	}
}
