package api

import (
	"encoding/json"
	"html"
	"net/http"
	"strconv"

	"github.com/urandom/readeef"
	"github.com/urandom/text-summary/summarize"
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
	identifier := ":article-id"

	patterns := map[string]webfw.MethodIdentifierTuple{
		prefix + "read/" + identifier + "/:value":     webfw.MethodIdentifierTuple{webfw.MethodPost, "read"},
		prefix + "favorite/" + identifier + "/:value": webfw.MethodIdentifierTuple{webfw.MethodPost, "favorite"},
		prefix + "formatter/" + identifier:            webfw.MethodIdentifierTuple{webfw.MethodGet, "formatter"},
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

		readeef.Debug.Printf("Invoking Article controller with action '%s', article id '%s'\n", action, params["article-id"])

		var articleId int64
		var article readeef.Article

		articleId, err = strconv.ParseInt(params["article-id"], 10, 64)

		if err == nil {
			article, err = db.GetFeedArticle(articleId, user)
		}

		resp := make(map[string]interface{})

		if err == nil {
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
				var formatting readeef.ArticleFormatting

				formatting, err = readeef.ArticleFormatter(webfw.GetConfig(c), con.config, article)
				if err != nil {
					break
				}

				s := summarize.NewFromString(formatting.Title, readeef.StripTags(formatting.Content))

				s.Language = formatting.Language
				keyPoints := s.KeyPoints()

				for i := range keyPoints {
					keyPoints[i] = html.UnescapeString(keyPoints[i])
				}

				resp["KeyPoints"] = keyPoints
				resp["Content"] = formatting.Content
				resp["TopImage"] = formatting.TopImage
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
