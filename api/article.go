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
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/content/search"
	"github.com/urandom/text-summary/summarize"
)

func getArticle(w http.ResponseWriter, r *http.Request) {
	if article, stop := articleFromRequest(w, r); stop {
		return
	} else {
		args{"article": article}.WriteJSON(w)
	}
}

type articleRepoType int

const (
	noRepoType articleRepoType = iota
	userRepoType
	favoriteRepoType
	popularRepoType
	tagRepoType
	feedRepoType
)

func getArticles(
	service repo.Service,
	repoType articleRepoType,
	subType articleRepoType,
	processors []content.ArticleProcessor,
	log readeef.Logger,
) http.HandlerFunc {
	repo := service.ArticleRepo()
	tagRepo := service.TagRepo()

	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		o, stop := articleQueryOptions(w, r)
		if stop {
			return
		}

		switch repoType {
		case favoriteRepoType:
			o = append(o, content.FavoriteOnly)
		case userRepoType:
		case popularRepoType:
			o = append(o, content.IncludeScores)
			o = append(o, content.HighScoredFirst)
			o = append(o, content.TimeRange(time.Now().AddDate(0, 0, -5), time.Now()))

			switch subType {
			case userRepoType:
			case tagRepoType:
				tag, stop := tagFromRequest(w, r)
				if stop {
					return
				}

				ids, err := tagRepo.FeedIDs(tag, user)
				if err != nil {
					error(w, log, "Error getting tag feed ids: %+v", err)
					return
				}

				o = append(o, content.FeedIDs(ids))
			case feedRepoType:
				feed, stop := feedFromRequest(w, r)
				if stop {
					return
				}

				o = append(o, content.FeedIDs([]content.FeedID{feed.ID}))
			default:
				http.Error(w, "Unknown article repository", http.StatusBadRequest)
				return
			}
		case tagRepoType:
			tag, stop := tagFromRequest(w, r)
			if stop {
				return
			}

			ids, err := tagRepo.FeedIDs(tag, user)
			if err != nil {
				error(w, log, "Error getting tag feed ids: %+v", err)
				return
			}

			o = append(o, content.FeedIDs(ids))
		case feedRepoType:
			feed, stop := feedFromRequest(w, r)
			if stop {
				return
			}

			o = append(o, content.FeedIDs([]content.FeedID{feed.ID}))
		default:
			http.Error(w, "Unknown article repository", http.StatusBadRequest)
			return
		}

		articles, err := repo.ForUser(user, o...)

		if err != nil {
			error(w, log, "Error getting articles: %+v", err)
			return
		}

		articles = content.ArticleProcessors(processors).Process(articles)

		args{"articles": articles}.WriteJSON(w)
	}
}

func articleSearch(
	service repo.Service,
	searchProvider content.SearchProvider,
	repoType articleRepoType,
	processors []content.ArticleProcessor,
	log readeef.Logger,
) http.HandlerFunc {
	repo := service.FeedRepo()
	tagRepo := service.TagRepo()

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

		switch repoType {
		case userRepoType:
		case feedRepoType:
			feed, stop := feedFromRequest(w, r)
			if stop {
				return
			}

			o = append(o, content.FeedIDs([]content.FeedID{feed.ID}))
		case tagRepoType:
			tag, stop := tagFromRequest(w, r)
			if stop {
				return
			}

			ids, err := tagRepo.FeedIDs(tag, user)
			if err != nil {
				error(w, log, "Error getting tag feed ids: %+v", err)
				return
			}

			o = append(o, content.FeedIDs(ids))
		default:
			http.Error(w, "Unknown repo type: "+repoType.String(), http.StatusBadRequest)
			return
		}

		articles, err := searchProvider.Search(req.Search, user, opts...)

		if err != nil {
			error(w, log, "Error searching for articles: %+v", err)
			return
		}

		articles = content.ArticleProcessors(processors).Process(articles)

		args{"articles": articles}.WriteJSON(w)
	}
}

func formatArticle(
	repo repo.Extract,
	extractor content.Extractor,
	processors []content.ArticleProcessors,
	log readeef.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		article, stop := articleFromRequest(w, r)
		if stop {
			return
		}

		extract, err := repo.Get(article.ID)

		if err != nil {
			if !content.IsNoContent(Err) {
				error(w, log, "Error getting article extract: %+v", err)
				return
			}

			if extract, err = extractor.Generate(article.Link); err != nil {
				http.Error(w, "Error getting article extract: "+err.Error(), http.StatusInternalServerError)
				return
			}

			extract.ArticleID = article.ID

			err := repo.Update(extract)
			if err != nil {
				error(w, log, "Error updating article extract: %+v", err)
				return
			}

		}

		if len(processors) > 0 {
			a := content.Article{Description: extract.Content}

			articles := []content.Article{a}

			if extract.TopImage != "" {
				articles = append(articles, content.Article{
					Description: fmt.Sprintf(`<img src="%s">`, extract.TopImage),
				})
			}

			articles = content.ArticleProcessors(processors).Process(articles)

			extract.Content = articles[0].Description

			if extract.TopImage != "" {
				content := articles[1].Description

				content = strings.Replace(content, `<img src="`, "", -1)
				i := strings.Index(content, `"`)
				content = content[:i]

				extract.TopImage = content
			}
		}

		s := summarize.NewFromString(extract.Title, search.StripTags(extract.Content))

		s.Language = extract.Language
		keyPoints := s.KeyPoints()

		for i := range keyPoints {
			keyPoints[i] = html.UnescapeString(keyPoints[i])
		}

		args{
			"keyPoints": keyPoints,
			"content":   extract.Content,
			"topImage":  extract.TopImage,
		}.WriteJSON(w)
	}
}

type articleState int

const (
	read articleState = iota
	favorite
)

func articleStateChange(
	repo repo.Article,
	state articleState,
	log readeef.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		article, stop := articleFromRequest(w, r)
		if stop {
			return
		}

		var value bool
		if stop = readJSON(w, r.Body, &value); stop {
			return
		}

		var previousState bool

		if state == read {
			previousState = article.Read
		} else {
			previousState = article.Favorite
		}

		if previousState != value {
			var err error
			if state == read {
				err = repo.Read(value, user, content.IDs([]content.ArticleID{article.ID}))
			} else {
				err = repo.Favor(value, user, content.IDs([]content.ArticleID{article.ID}))
			}

			if err != nil {
				error(w, log, "Error setting article "+state.String()+"state: %+v", err)
				return
			}
		}

		args{
			"success":      previousState != value,
			state.String(): value,
		}.WriteJSON(w)
	}
}

func articlesReadStateChange(
	service repo.Service,
	repoType articleRepoType,
	log readeef.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		var value bool
		if stop = readJSON(w, r.Body, &value); stop {
			return
		}

		o, stop := articleQueryOptions(w, r)
		if stop {
			return
		}

		switch repoType {
		case userRepoType:
		case favoriteRepoType:
			o = append(o, content.FavoriteOnly)
		case tagRepoType:
			tag, stop = tagFromRequest(w, r)
			if stop {
				return
			}

			ids, err := tagRepo.FeedIDs(tag, user)
			if err != nil {
				error(w, log, "Error getting tag feed ids: %+v", err)
				return
			}

			o = append(o, content.FeedIDs(ids))
		case feedRepoType:
			feed, stop = feedFromRequest(w, r)
			if stop {
				return
			}

			o = append(o, content.FeedIDs(content.FeedID{feed.ID}))
		default:
			http.Error(w, "Unknown type", http.StatusBadRequest)
			return
		}

		err = repo.Read(value, user, o...)

		if err != nil {
			error(w, log, "Error setting read state: %+v", err)
			return
		}

		args{"success": true}.WriteJSON(w)
	}
}

func articleQueryOptions(w http.ResponseWriter, r *http.Request) ([]content.QueryOpt, bool) {
	o := []content.QueryOpt{}

	query := r.URL.Query()

	var err error
	var limit, offset int
	if query.Get("limit") != "" {
		limit, err = strconv.Atoi(query.Get("limit"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return o, true
		}
	}

	if query.Get("offset") != "" {
		offset, err = strconv.Atoi(query.Get("offset"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return o, true
		}
	}

	if limit > 0 {
		o = append(o, content.Paging(limit, offset))
	}

	var minID, maxID int
	if query.Get("minID") != "" {
		minID, err = strconv.ParseInt(query.Get("minID"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return o, true
		}
	}

	if query.Get("maxID") != "" {
		maxID, err = strconv.ParseInt(query.Get("maxID"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return o, true
		}
	}

	if minID > 0 || maxID > 0 {
		o = append(o, content.IDRange(content.ArticleID(minID), content.ArticleID(maxID)))
	}

	if query.Get("unreadOnly") == "true" {
		o = append(o, content.UnreadOnly)
	}

	if query.Get("unreadFirst") == "true" {
		o = append(o, content.UnreadFirst)
	}

	if query.Get("olderFirst") == "true" {
		o = append(o, content.Sorting(content.SortByDate, content.AscendingOrder))
	} else {
		o = append(o, content.Sorting(content.SortByDate, content.DescendingOrder))
	}

	return o, false
}

func articleContext(repo repo.Article, processors []content.ArticleProcessor, log readeef.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, stop := userFromRequest(w, r)
			if stop {
				return
			}

			id, err := strconv.ParseInt(chi.URLParam(r, "articleID"), 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			articles, err := repo.ForUser(user, content.IDs([]content.ArticleID{id}))
			if err != nil {
				error(w, log, "Error getting article: %+v", err)
				return
			}

			if len(articles) == 0 {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}

			if r.Method == method.GET {
				articles = content.ArticleProcessors(processors).Process(articles)
			}

			ctx := context.WithValue(r.Context(), "article", articles[0])
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func articleFromRequest(w http.ResponseWriter, r *http.Request) (article content.UserArticle, stop bool) {
	var ok bool
	if article, ok = r.Context().Value("article").(content.UserArticle); ok {
		return article, false
	}

	http.Error(w, "Bad Request", http.StatusBadRequest)
	return nil, true
}
