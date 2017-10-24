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
	articlesLimit int,
	log log.Log,
) http.HandlerFunc {
	repo := service.ArticleRepo()
	tagRepo := service.TagRepo()

	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		o, stop := articleQueryOptions(w, r, articlesLimit)
		if stop {
			return
		}

		o = append(o, content.Filters(content.GetUserFilters(user)))

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

		if articles == nil {
			articles = []content.Article{}
		}
		args{"articles": articles}.WriteJSON(w)
	}
}

func articleSearch(
	service repo.Service,
	searchProvider search.Provider,
	repoType articleRepoType,
	processors []processor.Article,
	articlesLimit int,
	log log.Log,
) http.HandlerFunc {
	repo := service.TagRepo()

	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		if query == "" {
			http.Error(w, "No query provided", http.StatusBadRequest)
			return
		}

		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		o, stop := articleQueryOptions(w, r, articlesLimit)
		if stop {
			return
		}

		o = append(o, content.Filters(content.GetUserFilters(user)))

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

		if articles == nil {
			articles = []content.Article{}
		}
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

type articleStateEvent struct {
	user  content.User
	State string              `json:"state"`
	IDs   []content.ArticleID `json:"ids"`
	Value bool                `json:"value"`
}

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

		value := r.Method == http.MethodPost

		var previousState bool

		if state == read {
			previousState = article.Read
		} else {
			previousState = article.Favorite
		}

		if previousState != value {
			var err error
			ids := []content.ArticleID{article.ID}

			if state == read {
				err = repo.Read(value, user, content.IDs(ids))
			} else {
				err = repo.Favor(value, user, content.IDs(ids))
			}

			if err != nil {
				fatal(w, log, "Error setting article "+state.String()+"state: %+v", err)
				return
			}
		}

		args{
			"success":      true,
			state.String(): value,
		}.WriteJSON(w)
	}
}

func articlesReadStateChange(
	service repo.Service,
	repoType articleRepoType,
	articlesLimit int,
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

		o, stop := articleQueryOptions(w, r, articlesLimit)
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

func articleQueryOptions(w http.ResponseWriter, r *http.Request, articlesLimit int) ([]content.QueryOpt, bool) {
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

	if limit == 0 || limit > articlesLimit {
		limit = articlesLimit
	}

	o = append(o, content.Paging(limit, offset))

	var afterID, beforeID int64
	if query.Get("afterID") != "" {
		afterID, err = strconv.ParseInt(query.Get("afterID"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return o, true
		}
	}

	if query.Get("beforeID") != "" {
		beforeID, err = strconv.ParseInt(query.Get("beforeID"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return o, true
		}
	}

	if afterID > 0 || beforeID > 0 {
		o = append(o, content.IDRange(content.ArticleID(afterID), content.ArticleID(beforeID)))
	}

	var afterTime, beforeTime time.Time
	if query.Get("afterTime") != "" {
		seconds, err := strconv.ParseInt(query.Get("afterTime"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return o, true
		}
		afterTime = time.Unix(seconds, 0)
	}

	if query.Get("beforeTime") != "" {
		seconds, err := strconv.ParseInt(query.Get("beforeTime"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return o, true
		}
		beforeTime = time.Unix(seconds, 0)
	}

	if !afterTime.IsZero() || !beforeTime.IsZero() {
		o = append(o, content.TimeRange(afterTime, beforeTime))
	}

	if queryIDs, ok := query["id"]; ok {
		ids := make([]content.ArticleID, len(queryIDs))
		for i := range queryIDs {
			id, err := strconv.ParseInt(queryIDs[i], 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return o, true
			}
			ids[i] = content.ArticleID(id)
		}

		o = append(o, content.IDs(ids))
	}

	if _, ok := query["unreadOnly"]; ok {
		o = append(o, content.UnreadOnly)
	}

	if _, ok := query["unreadFirst"]; ok {
		o = append(o, content.UnreadFirst)
	}

	if _, ok := query["olderFirst"]; ok {
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
