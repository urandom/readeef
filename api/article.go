package api

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/urandom/handler/method"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base/search"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/text-summary/summarize"
)

func getArticle(w http.ResponseWriter, r *http.Request) {
	if article, stop := articleFromRequest(w, r); stop {
		return
	} else {
		args{"article": article}.WriteJSON(w)
	}
}

func formatArticle(extractor content.Extractor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		article, stop := articleFromRequest(w, r)
		if stop {
			return
		}

		extract := article.Extract()

		extractData := extract.Data()
		if extract.HasErr() {
			switch err := extract.Err(); err {
			case content.ErrNoContent:
				if extractData, err = extractor.Extract(article.Data().Link); err != nil {
					http.Error(w, "Error getting article extract: "+err.Error(), http.StatusInternalServerError)
					return
				}

				extractData.ArticleId = article.Data().Id
				extract.Data(extractData)
				extract.Update()
				if extract.HasErr() {
					http.Error(w, "Error updating article extract: "+extract.Err().Error(), http.StatusInternalServerError)
					return
				}
			default:
				http.Error(w, "Error getting article extract: "+err.Error(), http.StatusInternalServerError)
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

		args{
			"keyPoints": keyPoints,
			"content":   extractData.Content,
			"topImage":  extractData.TopImage,
		}.WriteJSON(w)
	}
}

type articleState int

const (
	read articleState = iota
	favorite
)

func articleStateChange(state articleState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		article, stop := articleFromRequest(w, r)
		if stop {
			return
		}

		var value bool
		if stop = readJSON(w, r.Body, &value); stop {
			return
		}

		in := article.Data()
		var previousState bool

		if state == read {
			previousState = in.Read
		} else {
			previousState = in.Favorite
		}

		if previousState != value {
			if state == read {
				article.Read(value)
			} else {
				article.Favorite(value)
			}

			if article.HasErr() {
				http.Error(w, "Error setting article "+state.String()+" state: "+article.Err().Error(), http.StatusInternalServerError)
				return
			}
		}

		args{
			"success":      previousState != value,
			state.String(): value,
		}.WriteJSON(w)
	}
}

type articleRepoType int

const (
	userRepoType articleRepoType = iota
	favoriteRepoType
	popularRepoType
	tagRepoType
	feedRepoType
)

func articlesReadStateChange(repoType articleRepoType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		o, stop := articleUpdateStateOptions(w, r)
		if stop {
			return
		}

		var ar content.ArticleRepo

		switch repoType {
		case userRepoType:
			ar = user
		case favoriteRepoType:
			ar = user
			o.FavoriteOnly = true
		case tagRepoType:
			ar, stop = tagFromRequest(w, r)
			if stop {
				return
			}
		case feedRepoType:
			ar, stop = feedFromRequest(w, r)
			if stop {
				return
			}
		default:
			http.Error(w, "Unknown type", http.StatusBadRequest)
			return
		}

		ar.ReadState(true, o)

		if e, ok := ar.(content.Error); ok && e.HasErr() {
			http.Error(w, "Error setting read state: "+e.Err().Error(), http.StatusInternalServerError)
			return
		}

		args{"success": true}.WriteJSON(w)
	}
}

func articleUpdateStateOptions(w http.ResponseWriter, r *http.Request) (data.ArticleUpdateStateOptions, bool) {
	o := data.ArticleUpdateStateOptions{}

	query := r.URL.Query()
	if query.Get("until") != "" {
		until, err := strconv.ParseInt(query.Get("until"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return o, true
		}

		o.BeforeDate = time.Unix(until, 0)
	}

	if query.Get("beforeArticle") != "" {
		before, err := strconv.ParseInt(query.Get("beforeArticle"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return o, true
		}

		o.BeforeId = data.ArticleId(before)
	}

	return o, false
}

func getArticles(repoType articleRepoType, subTypes ...articleRepoType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		o, stop := articleQueryOptions(w, r)
		if stop {
			return
		}

		var ar content.ArticleRepo

		switch repoType {
		case favoriteRepoType:
			o.FavoriteOnly = true
			fallthrough
		case userRepoType:
			ar = user
		case popularRepoType:
			o.IncludeScores = true
			o.HighScoredFirst = true
			o.BeforeDate = time.Now()
			o.AfterDate = time.Now().AddDate(0, 0, -5)

			if len(subTypes) > 0 {
				subType := subTypes[0]
				switch subType {
				case userRepoType:
					ar = user
				case tagRepoType:
					tag, stop := tagFromRequest(w, r)
					if stop {
						return
					}

					ar = tag
				case feedRepoType:
					feed, stop := feedFromRequest(w, r)
					if stop {
						return
					}

					ar = feed
				}
			}
		case tagRepoType:
			tag, stop := tagFromRequest(w, r)
			if stop {
				return
			}

			ar = tag
		case feedRepoType:
			feed, stop := feedFromRequest(w, r)
			if stop {
				return
			}

			ar = feed
		}

		if as, ok := ar.(content.ArticleSorting); ok {
			as.SortingByDate()
			if r.URL.Query().Get("olderFirst") == "true" {
				as.Order(data.AscendingOrder)
			} else {
				as.Order(data.DescendingOrder)
			}
		}

		if ar != nil {
			ua := ar.Articles(o)

			if e, ok := ar.(content.Error); ok && e.HasErr() {
				http.Error(w, "Error getting articles: "+e.Err().Error(), http.StatusInternalServerError)
			}

			args{"articles": ua}.WriteJSON(w)
		}

		http.Error(w, "Unknown article repository", http.StatusBadRequest)
	}
}

func articleSearch(searchProvider content.SearchProvider, repoType articleRepoType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := chi.URLParam(r, "*")
		if query == "" {
			http.Error(w, "No query provided", http.StatusBadRequest)
			return
		}

		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		o, stop := articleQueryOptions(w, r)
		if stop {
			return
		}

		searchProvider.SortingByDate()
		if r.URL.Query().Get("olderFirst") == "true" {
			searchProvider.Order(data.AscendingOrder)
		} else {
			searchProvider.Order(data.DescendingOrder)
		}

		var as content.ArticleSearch
		switch repoType {
		case userRepoType:
			as = user
		case feedRepoType:
			feed, stop := feedFromRequest(w, r)
			if stop {
				return
			}

			as = feed
		case tagRepoType:
			tag, stop := tagFromRequest(w, r)
			if stop {
				return
			}

			as = tag
		default:
			http.Error(w, "Unknown repo type: "+repoType.String(), http.StatusBadRequest)
			return
		}

		ua := as.Query(query, searchProvider, o.Limit, o.Offset)
		if e, ok := as.(content.Error); ok && e.HasErr() {
			http.Error(w, "Error while searching: "+e.Err().Error(), http.StatusInternalServerError)
			return
		}

		args{"articles": ua}.WriteJSON(w)
	}
}

func articleQueryOptions(w http.ResponseWriter, r *http.Request) (data.ArticleQueryOptions, bool) {
	o := data.ArticleQueryOptions{UnreadFirst: true}

	query := r.URL.Query()
	if query.Get("limit") != "" {
		limit, err := strconv.Atoi(query.Get("limit"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return o, true
		}

		o.Limit = limit
	}

	if query.Get("offset") != "" {
		offset, err := strconv.Atoi(query.Get("offset"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return o, true
		}

		o.Offset = offset
	}

	if query.Get("minID") != "" {
		minID, err := strconv.ParseInt(query.Get("minID"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return o, true
		}

		o.BeforeId = data.ArticleId(minID)
	}

	if query.Get("maxID") != "" {
		maxID, err := strconv.ParseInt(query.Get("maxID"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return o, true
		}

		o.AfterId = data.ArticleId(maxID)
	}

	if query.Get("unreadOnly") != "" {
		unreadOnly := query.Get("unreadOnly") == "true"

		o.UnreadOnly = unreadOnly
	}

	return o, false
}

func articleContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		id, err := strconv.ParseInt(chi.URLParam(r, "articleId"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		o := data.ArticleQueryOptions{}
		if r.Method == method.POST {
			o.SkipProcessors = true
		}

		article := user.ArticleById(data.ArticleId(id), o)
		if article.HasErr() {
			err := article.Err()
			if err == content.ErrNoContent {
				http.Error(w, "Not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		ctx := context.WithValue(r.Context(), "article", article)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func articleFromRequest(w http.ResponseWriter, r *http.Request) (article content.UserArticle, stop bool) {
	var ok bool
	if article, ok = r.Context().Value("article").(content.UserArticle); ok {
		return article, false
	}

	http.Error(w, "Bad Request", http.StatusBadRequest)
	return nil, true
}
