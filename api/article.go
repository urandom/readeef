package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/urandom/readeef"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
)

type Article struct {
	webfw.BaseController
	config readeef.Config
}

func NewArticle(config readeef.Config) Article {
	return Article{
		webfw.NewBaseController("", webfw.MethodAll, ""),
		config,
	}
}

type Readability struct {
	Content string
}

func (con Article) Patterns() map[string]webfw.MethodIdentifierTuple {
	prefix := "/v:version/article/"
	identifier := ":feed-id/:article-id"

	patterns := map[string]webfw.MethodIdentifierTuple{
		prefix + "read/" + identifier + "/:value":     webfw.MethodIdentifierTuple{webfw.MethodPost, "read"},
		prefix + "favorite/" + identifier + "/:value": webfw.MethodIdentifierTuple{webfw.MethodPost, "favorite"},
	}

	if con.config.ArticleFormatter.ReadabilityKey != "" {
		patterns[prefix+"formatter/"+identifier] = webfw.MethodIdentifierTuple{webfw.MethodGet, "formatter"}
	}

	return patterns
}

func (con Article) Handler(c context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		db := readeef.GetDB(c)
		user := readeef.GetUser(c, r)

		params := webfw.GetParams(c, r)
		action := webfw.GetMultiPatternIdentifier(c, r)

		readeef.Debug.Printf("Invoking Article controller with action '%s', feed id '%s' and article id '%s'\n", action, params["feed-id"], params["article-id"])

		var feedId, articleId int64
		feedId, err = strconv.ParseInt(params["feed-id"], 10, 64)

		var article readeef.Article
		if err == nil {
			articleId, err = strconv.ParseInt(params["article-id"], 10, 64)

			if err == nil {
				article, err = db.GetFeedArticle(feedId, articleId, user)
			}
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
			case "formatter":
				if con.config.ArticleFormatter.ReadabilityKey != "" {

					url := fmt.Sprintf("http://readability.com/api/content/v1/parser?url=%s&token=%s",
						url.QueryEscape(article.Link), con.config.ArticleFormatter.ReadabilityKey,
					)

					var response *http.Response
					var r Readability

					response, err = http.Get(url)

					if err != nil {
						break
					}

					defer response.Body.Close()
					dec := json.NewDecoder(response.Body)

					err = dec.Decode(&r)
					if err != nil {
						break
					}

					resp["Content"] = r.Content
				}
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
