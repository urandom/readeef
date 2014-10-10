package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/urandom/readeef"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
)

type Search struct {
	webfw.BaseController
	searchIndex readeef.SearchIndex
}

func NewSearch(searchIndex readeef.SearchIndex) Search {
	return Search{
		webfw.NewBaseController("/v:version/search/:query", webfw.MethodGet, ""),
		searchIndex,
	}
}

func (con Search) Handler(c context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		params := webfw.GetParams(c, r)
		query := params["query"]
		user := readeef.GetUser(c, r)

		resp := make(map[string]interface{})

		err = r.ParseForm()

		if err == nil {
			highlight := ""
			feedId := ""

			if vs := r.Form["highlight"]; len(vs) > 0 {
				highlight = vs[0]
			}

			if vs := r.Form["id"]; len(vs) > 0 {
				if vs[0] != "tag:__all__" && vs[0] != "__favorite__" {
					feedId = vs[0]
				}
			}

			var ids []int64
			if strings.HasPrefix(feedId, "tag:") {
				db := readeef.GetDB(c)

				var feeds []readeef.Feed

				feeds, err = db.GetUserTagFeeds(user, feedId[4:])

				if err == nil {
					for _, feed := range feeds {
						ids = append(ids, feed.Id)
					}
				}
			} else {
				if id, err := strconv.ParseInt(feedId, 10, 64); err == nil {
					ids = append(ids, id)
				}
			}

			var results []readeef.SearchResult

			results, err = con.searchIndex.Search(user, query, highlight, ids)

			if err == nil {
				resp["Articles"] = results
			}
		}

		var b []byte
		if err == nil {
			b, err = json.Marshal(resp)
		}
		if err != nil {
			webfw.GetLogger(c).Print(err)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(b)
	}
}

func (con Search) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}
