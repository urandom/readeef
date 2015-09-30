package api

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base/search"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/text-summary/summarize"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
)

type Article struct {
	config    readeef.Config
	extractor content.Extractor
}

func NewArticle(config readeef.Config, extractor content.Extractor) Article {
	return Article{config: config, extractor: extractor}
}

type Readability struct {
	Content string
}

type articleReadStateProcessor struct {
	Id    data.ArticleId `json:"id"`
	Value bool           `json:"value"`

	user content.User
}

type articleFavoriteStateProcessor struct {
	Id    data.ArticleId `json:"id"`
	Value bool           `json:"value"`

	user content.User
}

type formatArticleProcessor struct {
	Id data.ArticleId `json:"id"`

	user          content.User
	extractor     content.Extractor
	webfwConfig   webfw.Config
	readeefConfig readeef.Config
}

type getArticleProcessor struct {
	Id data.ArticleId `json:"id"`

	user content.User
}

func (con Article) Patterns() []webfw.MethodIdentifierTuple {
	prefix := "/v:version/article/:article-id/"

	return []webfw.MethodIdentifierTuple{
		webfw.MethodIdentifierTuple{"", webfw.MethodGet, "fetch"},
		webfw.MethodIdentifierTuple{prefix + "read/:value", webfw.MethodPost, "read"},
		webfw.MethodIdentifierTuple{prefix + "favorite/:value", webfw.MethodPost, "favorite"},
		webfw.MethodIdentifierTuple{prefix + "format", webfw.MethodGet, "format"},
	}
}

func (con Article) Handler(c context.Context) http.Handler {
	logger := webfw.GetLogger(c)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := readeef.GetUser(c, r)

		params := webfw.GetParams(c, r)
		action := webfw.GetMultiPatternIdentifier(c, r)

		logger.Infof("Invoking Article controller with action '%s', article id '%s'\n", action, params["article-id"])

		var articleId int64
		var resp responseError

		articleId, resp.err = strconv.ParseInt(params["article-id"], 10, 64)

		if resp.err == nil {
			id := data.ArticleId(articleId)
			switch action {
			case "fetch":
				resp = fetchArticle(user, id)
			case "read":
				resp = articleReadState(user, id, params["value"] == "true")
			case "favorite":
				resp = articleFavoriteState(user, id, params["value"] == "true")
			case "format":
				resp = formatArticle(user, id, con.extractor, webfw.GetConfig(c), con.config)
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

func (p articleReadStateProcessor) Process() responseError {
	return articleReadState(p.user, p.Id, p.Value)
}

func (p articleFavoriteStateProcessor) Process() responseError {
	return articleFavoriteState(p.user, p.Id, p.Value)
}

func (p formatArticleProcessor) Process() responseError {
	return formatArticle(p.user, p.Id, p.extractor, p.webfwConfig, p.readeefConfig)
}

func (p getArticleProcessor) Process() responseError {
	return fetchArticle(p.user, p.Id)
}

func fetchArticle(user content.User, id data.ArticleId) (resp responseError) {
	resp = newResponse()

	article := user.ArticleById(id)
	if user.HasErr() {
		resp.err = user.Err()
		return
	}

	resp.val["Article"] = article
	return
}

func articleReadState(user content.User, id data.ArticleId, read bool) (resp responseError) {
	resp = newResponse()

	article := user.ArticleById(id, data.ArticleQueryOptions{SkipProcessors: true})
	if user.HasErr() {
		resp.err = user.Err()
		return
	}

	in := article.Data()
	previouslyRead := in.Read

	if previouslyRead != read {
		article.Read(read)
		if article.HasErr() {
			resp.err = article.Err()
			return
		}
	}

	resp.val["Success"] = previouslyRead != read
	resp.val["Read"] = read
	resp.val["Id"] = in.Id
	return
}

func articleFavoriteState(user content.User, id data.ArticleId, favorite bool) (resp responseError) {
	resp = newResponse()

	article := user.ArticleById(id, data.ArticleQueryOptions{SkipProcessors: true})
	if user.HasErr() {
		resp.err = user.Err()
		return
	}

	in := article.Data()
	previouslyFavorite := in.Favorite

	if previouslyFavorite != favorite {
		article.Favorite(favorite)
		if article.HasErr() {
			resp.err = article.Err()
			return
		}
	}

	resp.val["Success"] = previouslyFavorite != favorite
	resp.val["Favorite"] = favorite
	resp.val["Id"] = in.Id
	return
}

func formatArticle(user content.User, id data.ArticleId, extractor content.Extractor, webfwConfig webfw.Config, readeefConfig readeef.Config) (resp responseError) {
	resp = newResponse()

	article := user.ArticleById(id)
	if user.HasErr() {
		resp.err = user.Err()
		return
	}

	extract := article.Extract()
	if article.HasErr() {
		resp.err = article.Err()
		return
	}

	extractData := extract.Data()
	if extract.HasErr() {
		switch err := extract.Err(); err {
		case content.ErrNoContent:
			if extractor == nil {
				resp.err = fmt.Errorf("Error formatting article: A valid extractor is reequired")
				return
			}

			extractData, resp.err = extractor.Extract(article.Data().Link)
			if resp.err != nil {
				return
			}

			extractData.ArticleId = article.Data().Id
			extract.Data(extractData)
			extract.Update()
			if extract.HasErr() {
				resp.err = extract.Err()
				return
			}
		default:
			resp.err = err
			return
		}
	}

	processors := user.Repo().ArticleProcessors()
	if len(processors) > 0 {
		a := user.Repo().UserArticle(user)
		a.Data(data.Article{Description: extractData.Content})

		ua := []content.UserArticle{a}

		if extractData.TopImage != "" {
			a = user.Repo().UserArticle(user)
			a.Data(data.Article{
				Description: fmt.Sprintf(`<img src="%s">`, extractData.TopImage),
			})

			ua = append(ua, a)
		}

		for _, p := range processors {
			ua = p.ProcessArticles(ua)
		}

		extractData.Content = ua[0].Data().Description

		if extractData.TopImage != "" {
			content := ua[1].Data().Description

			content = strings.Replace(content, `<img src="`, "", -1)
			i := strings.Index(content, `"`)
			content = content[:i]

			extractData.TopImage = content
		}
	}

	s := summarize.NewFromString(extractData.Title, search.StripTags(extractData.Content))

	s.Language = extractData.Language
	keyPoints := s.KeyPoints()

	for i := range keyPoints {
		keyPoints[i] = html.UnescapeString(keyPoints[i])
	}

	resp.val["KeyPoints"] = keyPoints
	resp.val["Content"] = extractData.Content
	resp.val["TopImage"] = extractData.TopImage
	resp.val["Id"] = id
	return
}
