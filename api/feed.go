package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"readeef"
	"readeef/parser"
	"strings"

	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
)

type Feed struct {
	webfw.BaseController
}

func NewFeed() Feed {
	return Feed{
		webfw.NewBaseController("/v:version/feed/*action", webfw.MethodAll, ""),
	}
}

func (con Feed) Handler(c context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		db := readeef.GetDB(c)
		user := readeef.GetUser(c, r)

		actionParam := webfw.GetParams(c, r)
		parts := strings.SplitN(actionParam["action"], "/", 2)
		action := parts[0]

		switch action {
		case "all":
			var feeds []readeef.Feed
			feeds, err = db.GetUserFeeds(user)

			if err != nil {
				break
			}

			type feed struct {
				Title       string
				Description string
				Link        string
				Image       parser.Image
			}
			type response struct {
				Feeds []feed
			}

			resp := response{}
			for _, f := range feeds {
				resp.Feeds = append(resp.Feeds, feed{f.Title, f.Description, f.Link, f.Image})
			}

			var b []byte
			b, err = json.Marshal(resp)
			if err != nil {
				break
			}

			w.Write(b)
		default:
			err = errors.New("Unknown action " + action)
		}

		if err != nil {
			webfw.GetLogger(c).Print(err)

			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (con Feed) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}
