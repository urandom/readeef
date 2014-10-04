package api

import (
	"encoding/json"
	"net/http"
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

		resp := make(map[string]interface{})

		err = r.ParseForm()

		if err == nil {
			highlight := "html"
			if vs := r.Form["highlight"]; len(vs) > 0 {
				highlight = vs[0]
			}

			var results []readeef.SearchResult

			results, err = con.searchIndex.Search(query, highlight)

			if err == nil {
				resp["Results"] = results
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
