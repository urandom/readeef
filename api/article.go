package api

import (
	"encoding/json"
	"net/http"
	"readeef"
	"strconv"

	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
)

type Article struct {
	webfw.BaseController
}

func NewArticle() Article {
	return Article{
		webfw.NewBaseController("/v:version/article/:action/:feedId/:articleId/:value", webfw.MethodAll, ""),
	}
}

func (con Article) Handler(c context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		db := readeef.GetDB(c)
		user := readeef.GetUser(c, r)

		params := webfw.GetParams(c, r)
		action := params["action"]

		readeef.Debug.Printf("Invoking Article controller with action '%s', feed id '%s' and article id '%s'\n", action, params["feedId"], params["articleId"])

		var feedId int64
		feedId, err = strconv.ParseInt(params["feedId"], 10, 64)

		var article readeef.Article
		if err == nil {
			article, err = db.GetFeedArticle(feedId, params["articleId"], user)
		}

		if err == nil {

			switch action {
			case "read":
				read := params["value"] == "true"
				previouslyRead := article.Read

				if !previouslyRead {
					err = db.MarkUserArticlesAsRead(user, []readeef.Article{article}, read)

					if err != nil {
						break
					}
				}

				type response struct {
					Success bool
					Read    bool
				}

				resp := response{Success: !previouslyRead, Read: read}

				var b []byte
				b, err = json.Marshal(resp)
				if err != nil {
					break
				}

				w.Write(b)
			case "favorite":
				favorite := params["value"] == "true"
				previouslyFavorite := article.Favorite

				if !previouslyFavorite {
					err = db.MarkUserArticlesAsFavorite(user, []readeef.Article{article}, favorite)

					if err != nil {
						break
					}
				}

				type response struct {
					Success  bool
					Favorite bool
				}

				resp := response{Success: !previouslyFavorite, Favorite: favorite}

				var b []byte
				b, err = json.Marshal(resp)
				if err != nil {
					break
				}

				w.Write(b)
			}
		}

		if err != nil {
			webfw.GetLogger(c).Print(err)

			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (con Article) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}
