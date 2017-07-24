package ttrss

import (
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base/search"
	"github.com/urandom/readeef/content/data"
)

type headlinesHeaderContent []interface{}
type headlinesContent []headline

type headline struct {
	Id        data.ArticleId `json:"id"`
	Unread    bool           `json:"unread"`
	Marked    bool           `json:"marked"`
	Updated   int64          `json:"updated"`
	IsUpdated bool           `json:"is_updated"`
	Title     string         `json:"title"`
	Link      string         `json:"link"`
	FeedId    string         `json:"feed_id"`
	Author    string         `json:"author"`
	Excerpt   string         `json:"excerpt,omitempty"`
	Content   string         `json:"content,omitempty"`
	FeedTitle string         `json:"feed_title"`

	Tags   []string `json:"tags,omitempty"`
	Labels []string `json:"labels,omitempty"`
}

type headlinesHeader struct {
	Id      data.FeedId    `json:"id"`
	FirstId data.ArticleId `json:"first_id"`
	IsCat   bool           `json:"is_cat"`
}

type articlesContent []article

type article struct {
	Id        string `json:"id"`
	Title     string `json:"title"`
	Link      string `json:"link"`
	Unread    bool   `json:"unread"`
	Marked    bool   `json:"marked"`
	Author    string `json:"author"`
	Updated   int64  `json:"updated"`
	Content   string `json:"content,omitempty"`
	FeedId    string `json:"feed_id"`
	FeedTitle string `json:"feed_title"`

	Labels []string `json:"labels,omitempty"`
}

func registerArticleActions(searchProvider content.SearchProvider) {
	actions["getHeadlines"] = func(req request, user content.User) (interface{}, error) {
		return getHeadlines(req, user, searchProvider)
	}
	actions["updateArticle"] = updateArticle
	actions["getArticle"] = getArticle
}

func getHeadlines(req request, user content.User, searchProvider content.SearchProvider) (interface{}, error) {
	if req.FeedId == 0 {
		return nil, errors.WithStack(newErr("no feed id", "INCORRECT_USAGE"))
	}

	limit := req.Limit
	if limit == 0 {
		limit = 200
	}

	var articles []content.UserArticle
	var articleRepo content.ArticleRepo
	var feedTitle string
	firstId := data.ArticleId(0)
	o := data.ArticleQueryOptions{Limit: limit, Offset: req.Skip, UnreadFirst: true, SkipSessionProcessors: true}

	if req.IsCat {
		if req.FeedId == CAT_UNCATEGORIZED {
			setupSorting(req, user)
			articleRepo = user
			o.UntaggedOnly = true
			feedTitle = "Uncategorized"
		} else if req.FeedId > 0 {
			t := user.TagById(data.TagId(req.FeedId))
			if t.HasErr() {
				return nil, errors.Wrapf(t.Err(), "getting user %s tag %d", user.Data().Login, req.FeedId)
			}

			setupSorting(req, t)
			articleRepo = t
			feedTitle = string(t.Data().Value)
		}
	} else {
		if req.FeedId == FAVORITE_ID {
			setupSorting(req, user)
			o.FavoriteOnly = true
			articleRepo = user
			feedTitle = "Starred articles"
		} else if req.FeedId == FRESH_ID {
			setupSorting(req, user)
			o.AfterDate = time.Now().Add(FRESH_DURATION)
			articleRepo = user
			feedTitle = "Fresh articles"
		} else if req.FeedId == ALL_ID {
			setupSorting(req, user)
			articleRepo = user
			feedTitle = "All articles"
		} else if req.FeedId > 0 {
			feed := user.FeedById(req.FeedId)

			if feed.HasErr() {
				return nil, errors.Wrapf(feed.Err(), "getting user %s feed %d", user.Data().Login, req.FeedId)
			}

			setupSorting(req, feed)
			articleRepo = feed
			feedTitle = feed.Data().Title
		}
	}

	if req.SinceId > 0 {
		o.AfterId = req.SinceId
	}

	if articleRepo != nil {
		if req.Search != "" {
			if searchProvider != nil {
				if as, ok := articleRepo.(content.ArticleSearch); ok {
					articles = as.Query(req.Search, searchProvider, limit, req.Skip)
				}
			}
		} else {
			var skip bool

			switch req.ViewMode {
			case "all_articles":
			case "adaptive":
			case "unread":
				o.UnreadOnly = true
			case "marked":
				o.FavoriteOnly = true
			default:
				skip = true
			}

			if !skip {
				articles = articleRepo.Articles(o)
			}
		}
	}

	if e, ok := articleRepo.(content.Error); ok {
		if e.HasErr() {
			return nil, errors.Wrap(e.Err(), "getting articles")
		}
	}

	if len(articles) > 0 {
		firstId = articles[0].Data().Id
	}

	headlines := headlinesFromArticles(articles, feedTitle, req.ShowContent, req.ShowExcerpt)
	if req.IncludeHeader {
		header := headlinesHeader{Id: req.FeedId, FirstId: firstId, IsCat: req.IsCat}
		hContent := headlinesHeaderContent{}

		hContent = append(hContent, header)
		hContent = append(hContent, headlines)

		return hContent, nil
	} else {
		return headlines, nil
	}
}

func updateArticle(req request, user content.User) (interface{}, error) {
	articles := user.ArticlesById(req.ArticleIds, data.ArticleQueryOptions{SkipSessionProcessors: true})
	updateCount := int64(0)

	if req.Field != 0 && req.Field != 2 {
		return nil, errors.Errorf("Unknown field %d", req.Field)
	}

	for _, a := range articles {
		d := a.Data()
		updated := false

		switch req.Field {
		case 0:
			switch req.Mode {
			case 0:
				if d.Favorite {
					updated = true
					d.Favorite = false
				}
			case 1:
				if !d.Favorite {
					updated = true
					d.Favorite = true
				}
			case 2:
				updated = true
				d.Favorite = !d.Favorite
			}
			if updated {
				a.Favorite(d.Favorite)
			}
		case 2:
			switch req.Mode {
			case 0:
				if !d.Read {
					updated = true
					d.Read = true
				}
			case 1:
				if d.Read {
					updated = true
					d.Read = false
				}
			case 2:
				updated = true
				d.Read = !d.Read
			}
			if updated {
				a.Read(d.Read)
			}
		}

		if updated {
			if a.HasErr() {
				return nil, errors.Wrapf(a.Err(), "marking article %d as %d:%d", a.Data().Id, req.Field, req.Mode)
			}

			updateCount++
		}
	}

	return genericContent{Status: "OK", Updated: updateCount}, nil
}

func getArticle(req request, user content.User) (interface{}, error) {
	articles := user.ArticlesById(req.ArticleId, data.ArticleQueryOptions{SkipSessionProcessors: true})
	if user.HasErr() {
		return nil, errors.Wrap(user.Err(), "getting user articles")
	}

	feedTitles := map[data.FeedId]string{}

	for _, a := range articles {
		d := a.Data()
		if _, ok := feedTitles[d.FeedId]; !ok {
			f := user.Repo().FeedById(d.FeedId)
			if f.HasErr() {
				return nil, errors.Wrapf(f.Err(), "getting feed by id %d", d.FeedId)
			}

			feedTitles[d.FeedId] = f.Data().Title
		}
	}

	cContent := articlesContent{}

	for _, a := range articles {
		d := a.Data()
		title := feedTitles[d.FeedId]
		h := article{
			Id:        strconv.FormatInt(int64(d.Id), 10),
			Unread:    !d.Read,
			Marked:    d.Favorite,
			Updated:   d.Date.Unix(),
			Title:     d.Title,
			Link:      d.Link,
			FeedId:    strconv.FormatInt(int64(d.FeedId), 10),
			FeedTitle: title,
			Content:   d.Description,
		}

		cContent = append(cContent, h)
	}

	return cContent, nil
}

func setupSorting(req request, sorting content.ArticleSorting) {
	switch req.OrderBy {
	case "date_reverse":
		sorting.SortingByDate()
		sorting.Order(data.AscendingOrder)
	default:
		sorting.SortingByDate()
		sorting.Order(data.DescendingOrder)
	}
}

func headlinesFromArticles(articles []content.UserArticle, feedTitle string, content, excerpt bool) headlinesContent {
	c := headlinesContent{}
	for _, a := range articles {
		d := a.Data()
		title := feedTitle
		h := headline{
			Id:        d.Id,
			Unread:    !d.Read,
			Marked:    d.Favorite,
			Updated:   d.Date.Unix(),
			IsUpdated: !d.Read,
			Title:     d.Title,
			Link:      d.Link,
			FeedId:    strconv.FormatInt(int64(d.FeedId), 10),
			FeedTitle: title,
		}

		if content {
			h.Content = d.Description
		}

		if excerpt {
			excerpt := search.StripTags(d.Description)
			if len(excerpt) > 100 {
				excerpt = excerpt[:100]
			}

			h.Excerpt = excerpt
		}

		c = append(c, h)
	}

	return c
}
