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
	config readeef.Config
}

func NewArticle(config readeef.Config) Article {
	return Article{config}
}

type Readability struct {
	Content string
}

func (con Article) Patterns() []webfw.MethodIdentifierTuple {
	prefix := "/v:version/article/:article-id/"

	return []webfw.MethodIdentifierTuple{
		webfw.MethodIdentifierTuple{prefix + "read/:value", webfw.MethodPost, "read"},
		webfw.MethodIdentifierTuple{prefix + "favorite/:value", webfw.MethodPost, "favorite"},
		webfw.MethodIdentifierTuple{prefix + "format", webfw.MethodGet, "format"},
	}
}

func (con Article) Handler(c context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db := readeef.GetDB(c)
		user := readeef.GetUser(c, r)

		params := webfw.GetParams(c, r)
		action := webfw.GetMultiPatternIdentifier(c, r)

		readeef.Debug.Printf("Invoking Article controller with action '%s', article id '%s'\n", action, params["article-id"])

		var articleId int64
		var resp responseError

		articleId, resp.err = strconv.ParseInt(params["article-id"], 10, 64)

		if resp.err == nil {
			switch action {
			case "read":
				resp = markArticleAsRead(db, user, articleId, params["value"] == "true")
			case "favorite":
				resp = markArticleAsFavorite(db, user, articleId, params["value"] == "true")
			case "format":
				resp = formatArticle(db, user, articleId, webfw.GetConfig(c), con.config)
			}
		}

		var b []byte
		if resp.err == nil {
			b, resp.err = json.Marshal(resp.val)
		}

		if resp.err == nil {
			w.Write(b)
		} else {
			webfw.GetLogger(c).Print(resp.err)

			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func (con Article) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}

func getArticle(db readeef.DB, user readeef.User, id int64) (article readeef.Article, err error) {
	article, err = db.GetFeedArticle(id, user)
	return
}

func markArticleAsRead(db readeef.DB, user readeef.User, id int64, read bool) (resp responseError) {
	resp = newResponse()

	var article readeef.Article
	if article, resp.err = getArticle(db, user, id); resp.err != nil {
		return
	}

	previouslyRead := article.Read

	if previouslyRead != read {
		if resp.err = db.MarkUserArticlesAsRead(user, []readeef.Article{article}, read); resp.err != nil {
			return
		}
	}

	resp.val["Success"] = previouslyRead != read
	resp.val["Read"] = read
	resp.val["ArticleId"] = article.Id
	return
}

func markArticleAsFavorite(db readeef.DB, user readeef.User, id int64, favorite bool) (resp responseError) {
	resp = newResponse()

	var article readeef.Article
	if article, resp.err = getArticle(db, user, id); resp.err != nil {
		return
	}

	previouslyFavorite := article.Favorite

	if previouslyFavorite != favorite {
		if resp.err = db.MarkUserArticlesAsFavorite(user, []readeef.Article{article}, favorite); resp.err != nil {
			return
		}
	}

	resp.val["Success"] = previouslyFavorite != favorite
	resp.val["Favorite"] = favorite
	resp.val["ArticleId"] = article.Id
	return
}

func formatArticle(db readeef.DB, user readeef.User, id int64, webfwConfig webfw.Config, readeefConfig readeef.Config) (resp responseError) {
	resp = newResponse()

	var article readeef.Article
	if article, resp.err = getArticle(db, user, id); resp.err != nil {
		return
	}

	var formatting readeef.ArticleFormatting

	if formatting, resp.err = readeef.ArticleFormatter(webfwConfig, readeefConfig, article); resp.err != nil {
		return
	}

	s := summarize.NewFromString(formatting.Title, readeef.StripTags(formatting.Content))

	s.Language = formatting.Language
	keyPoints := s.KeyPoints()

	for i := range keyPoints {
		keyPoints[i] = html.UnescapeString(keyPoints[i])
	}

	resp.val["KeyPoints"] = keyPoints
	resp.val["Content"] = formatting.Content
	resp.val["TopImage"] = formatting.TopImage
	resp.val["ArticleId"] = article.Id
	return
}
