package api

import (
	"context"
	"html"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/extract"
	"github.com/urandom/readeef/content/processor"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/content/search"
	"github.com/urandom/readeef/log"
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
	processors []processor.Article,
	log log.Log,
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
					fatal(w, log, "Error getting tag feed ids: %+v", err)
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
				fatal(w, log, "Error getting tag feed ids: %+v", err)
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
			fatal(w, log, "Error getting articles: %+v", err)
			return
		}

		articles = processor.Articles(processors).Process(articles)

		args{"articles": articles}.WriteJSON(w)
	}
}

func articleSearch(
	repo repo.Tag,
	searchProvider search.Provider,
	repoType articleRepoType,
	processors []processor.Article,
	log log.Log,
) http.HandlerFunc {
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

			ids, err := repo.FeedIDs(tag, user)
			if err != nil {
				fatal(w, log, "Error getting tag feed ids: %+v", err)
				return
			}

			o = append(o, content.FeedIDs(ids))
		default:
			http.Error(w, "Unknown repo type: "+repoType.String(), http.StatusBadRequest)
			return
		}

		articles, err := searchProvider.Search(query, user, o...)

		if err != nil {
			fatal(w, log, "Error searching for articles: %+v", err)
			return
		}

		articles = processor.Articles(processors).Process(articles)

		args{"articles": articles}.WriteJSON(w)
	}
}

func formatArticle(
	repo repo.Extract,
	extractor extract.Generator,
	processors []processor.Article,
	log log.Log,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, stop := userFromRequest(w, r)
		if stop {
			return
		}

		article, stop := articleFromRequest(w, r)
		if stop {
			return
		}

		extract, err := extract.Get(article, repo, extractor, processors)
		if err != nil {
			fatal(w, log, "Error getting article extract: %+v", err)
			return
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
	log log.Log,
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
				fatal(w, log, "Error setting article "+state.String()+"state: %+v", err)
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
	log log.Log,
) http.HandlerFunc {
	articleRepo := service.ArticleRepo()
	tagRepo := service.TagRepo()

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
			tag, stop := tagFromRequest(w, r)
			if stop {
				return
			}

			ids, err := tagRepo.FeedIDs(tag, user)
			if err != nil {
				fatal(w, log, "Error getting tag feed ids: %+v", err)
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
			http.Error(w, "Unknown type", http.StatusBadRequest)
			return
		}

		if err := articleRepo.Read(value, user, o...); err != nil {
			fatal(w, log, "Error setting read state: %+v", err)
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

	if limit == 0 {
		limit = 200
	}

	o = append(o, content.Paging(limit, offset))

	var minID, maxID int64
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

func articleContext(repo repo.Article, processors []processor.Article, log log.Log) func(http.Handler) http.Handler {
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

			articles, err := repo.ForUser(user, content.IDs([]content.ArticleID{content.ArticleID(id)}))
			if err != nil {
				fatal(w, log, "Error getting article: %+v", err)
				return
			}

			if len(articles) == 0 {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}

			if r.Method == "GET" {
				articles = processor.Articles(processors).Process(articles)
			}

			ctx := context.WithValue(r.Context(), "article", articles[0])
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func articleFromRequest(w http.ResponseWriter, r *http.Request) (article content.Article, stop bool) {
	var ok bool
	if article, ok = r.Context().Value("article").(content.Article); ok {
		return article, false
	}

	http.Error(w, "Bad Request", http.StatusBadRequest)
	return content.Article{}, true
}
