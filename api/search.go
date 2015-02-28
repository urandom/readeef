package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/info"
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

	user content.User
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

			resp = search(user, con.searchIndex, query, highlight, feedId)
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
	return search(p.user, p.si, p.Query, p.Highlight, p.Id)
}

func search(user content.User, searchIndex readeef.SearchIndex, query, highlight, feedId string) (resp responseError) {
	resp = newResponse()

	if strings.HasPrefix(feedId, "tag:") {
		tag := user.Repo().Tag(user)
		tag.Value(info.TagValue(feedId[4:]))

		tag.Highlight(highlight)
		resp.val["Articles"], resp.err = tag.Query(query, searchIndex.Index), tag.Err()
	} else {
		if id, err := strconv.ParseInt(feedId, 10, 64); err == nil {
			f := user.FeedById(info.FeedId(id))
			resp.val["Articles"], resp.err = f.Query(query, searchIndex.Index), f.Err()
		} else {
			user.Highlight(highlight)
			resp.val["Articles"], resp.err = user.Query(query, searchIndex.Index), user.Err()
		}
	}

	return
}
