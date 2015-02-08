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
	webfw.BasePatternController
	searchIndex readeef.SearchIndex
}

type searchProcessor struct {
	Query     string `json:"query"`
	Highlight string `json:"highlight"`
	Id        string `json:"id"`

	db   readeef.DB
	user readeef.User
	si   readeef.SearchIndex
}

func NewSearch(searchIndex readeef.SearchIndex) Search {
	return Search{
		webfw.NewBasePatternController("/v:version/search/:query", webfw.MethodGet, ""),
		searchIndex,
	}
}

func (con Search) Handler(c context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db := readeef.GetDB(c)
		user := readeef.GetUser(c, r)
		params := webfw.GetParams(c, r)
		query := params["query"]

		var resp responseError

		if resp.err = r.ParseForm(); resp.err == nil {
			highlight := ""
			feedId := ""

			if vs := r.Form["highlight"]; len(vs) > 0 {
				highlight = vs[0]
			}

			if vs := r.Form["id"]; len(vs) > 0 {
				if vs[0] != "all" && vs[0] != "favorite" {
					feedId = vs[0]
				}
			}

			resp = search(db, user, con.searchIndex, query, highlight, feedId)
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

func (con Search) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}

func (p searchProcessor) Process() responseError {
	return search(p.db, p.user, p.si, p.Query, p.Highlight, p.Id)
}

func search(db readeef.DB, user readeef.User, searchIndex readeef.SearchIndex, query, highlight, feedId string) (resp responseError) {
	resp = newResponse()
	var ids []int64

	if feedId != "" {
		if strings.HasPrefix(feedId, "tag:") {
			var feeds []readeef.Feed
			if feeds, resp.err = db.GetUserTagFeeds(user, feedId[4:]); resp.err != nil {
				return
			}

			for _, feed := range feeds {
				ids = append(ids, feed.Id)
			}
		} else {
			if id, err := strconv.ParseInt(feedId, 10, 64); err == nil {
				ids = append(ids, id)
			}
		}
	}

	var results []readeef.SearchResult
	if results, resp.err = searchIndex.Search(user, query, highlight, ids); resp.err != nil {
		return
	}

	resp.val["Articles"] = results
	return
}
