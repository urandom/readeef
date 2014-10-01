package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/urandom/readeef"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
)

type Article struct {
	webfw.BaseController
}

func NewArticle() Article {
	return Article{
		webfw.NewBaseController("", webfw.MethodAll, ""),
	}
}

func (con Article) Patterns() map[string]webfw.MethodIdentifierTuple {
	prefix := "/v:version/article/"

	return map[string]webfw.MethodIdentifierTuple{
		prefix + "read/:feed-id/:article-id/:value":     webfw.MethodIdentifierTuple{webfw.MethodPost, "read"},
		prefix + "favorite/:feed-id/:article-id/:value": webfw.MethodIdentifierTuple{webfw.MethodPost, "favorite"},
	}
}

func (con Article) Handler(c context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		db := readeef.GetDB(c)
		user := readeef.GetUser(c, r)

		params := webfw.GetParams(c, r)
		action := webfw.GetMultiPatternIdentifier(c, r)

		readeef.Debug.Printf("Invoking Article controller with action '%s', feed id '%s' and article id '%s'\n", action, params["feed-id"], params["article-id"])

		var feedId int64
		feedId, err = strconv.ParseInt(params["feed-id"], 10, 64)

		var article readeef.Article
		if err == nil {
			article, err = db.GetFeedArticle(feedId, params["article-id"], user)
		}

		resp := make(map[string]interface{})

		if err == nil {
			resp["Id"] = feedId
			resp["ArticleId"] = article.Id

			switch action {
			case "read":
				read := params["value"] == "true"
				previouslyRead := article.Read

				if previouslyRead != read {
					err = db.MarkUserArticlesAsRead(user, []readeef.Article{article}, read)

					if err != nil {
						break
					}
				}

				resp["Success"] = previouslyRead != read
				resp["Read"] = read
			case "favorite":
				favorite := params["value"] == "true"
				previouslyFavorite := article.Favorite

				if previouslyFavorite != favorite {
					err = db.MarkUserArticlesAsFavorite(user, []readeef.Article{article}, favorite)

					if err != nil {
						break
					}
				}

				resp["Success"] = previouslyFavorite != favorite
				resp["Favorite"] = favorite
			}
		}

		var b []byte
		if err == nil {
			b, err = json.Marshal(resp)
		}

		if err == nil {
			w.Write(b)
		} else {
			webfw.GetLogger(c).Print(err)

			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (con Article) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}
