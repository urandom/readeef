package ttrss

import (
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/processor"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/content/search"
)

type headlinesHeaderContent []interface{}
type headlinesContent []headline

type headline struct {
	Id        content.ArticleID `json:"id"`
	Unread    bool              `json:"unread"`
	Marked    bool              `json:"marked"`
	Updated   int64             `json:"updated"`
	IsUpdated bool              `json:"is_updated"`
	Title     string            `json:"title"`
	Link      string            `json:"link"`
	FeedId    string            `json:"feed_id"`
	Author    string            `json:"author"`
	Excerpt   string            `json:"excerpt,omitempty"`
	Content   string            `json:"content,omitempty"`
	FeedTitle string            `json:"feed_title"`

	Tags   []string `json:"tags,omitempty"`
	Labels []string `json:"labels,omitempty"`
}

type headlinesHeader struct {
	Id      content.FeedID    `json:"id"`
	FirstId content.ArticleID `json:"first_id"`
	IsCat   bool              `json:"is_cat"`
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

func registerArticleActions(searchProvider search.Provider, processors []processor.Article) {
	actions["getHeadlines"] = func(req request, user content.User, service repo.Service) (interface{}, error) {
		return getHeadlines(req, user, service, searchProvider, processors)
	}
	actions["updateArticle"] = updateArticle
	actions["getArticle"] = func(req request, user content.User, service repo.Service) (interface{}, error) {
		return getArticle(req, user, service, processors)
	}
}

func getHeadlines(
	req request,
	user content.User,
	service repo.Service,
	searchProvider search.Provider,
	processors []processor.Article,
) (interface{}, error) {
	if req.FeedId == 0 {
		return nil, errors.WithStack(newErr("no feed id", "INCORRECT_USAGE"))
	}

	limit := req.Limit
	if limit == 0 {
		limit = 200
	}

	var feedTitle string
	var firstID content.ArticleID
	opts := []content.QueryOpt{
		content.Paging(limit, req.Skip), content.UnreadFirst,
		content.Filters(content.GetUserFilters(user)),
	}

	switch req.OrderBy {
	case "date_reverse":
		opts = append(opts, content.Sorting(content.SortByDate, content.AscendingOrder))
	default:
		opts = append(opts, content.Sorting(content.SortByDate, content.DescendingOrder))
	}

	var feedGenerator func() ([]content.Feed, error)

	if req.SinceId > 0 {
		opts = append(opts, content.IDRange(0, req.SinceId))
	}

	if req.IsCat {
		if req.FeedId == CAT_UNCATEGORIZED {
			opts = append(opts, content.UntaggedOnly)

			feedTitle = "Uncategorized"
		} else if req.FeedId > 0 {
			tag, err := service.TagRepo().Get(content.TagID(req.FeedId), user)
			if err != nil {
				return nil, errors.WithMessage(err, "getting tag for user")
			}

			feedGenerator = func() ([]content.Feed, error) {
				return service.FeedRepo().ForTag(tag, user)
			}

			feedTitle = string(tag.Value)
		}
	} else {
		if req.FeedId == FAVORITE_ID {
			opts = append(opts, content.FavoriteOnly)
			feedTitle = "Starred articles"
		} else if req.FeedId == FRESH_ID {
			opts = append(opts, content.TimeRange(time.Now().Add(FRESH_DURATION), time.Time{}))
			feedTitle = "Fresh articles"
		} else if req.FeedId == ALL_ID {
			feedTitle = "All articles"
		} else if req.FeedId > 0 {
			feed, err := service.FeedRepo().Get(req.FeedId, user)
			if err != nil {
				return nil, errors.WithMessage(err, "getting user feed")
			}

			feedGenerator = func() ([]content.Feed, error) {
				return []content.Feed{feed}, nil
			}

			feedTitle = feed.Title
		}
	}

	if feedGenerator != nil {
		feeds, err := feedGenerator()
		if err != nil {
			return nil, errors.WithMessage(err, "getting feeds")
		}

		ids := make([]content.FeedID, len(feeds))
		for i := range feeds {
			ids[i] = feeds[i].ID
		}

		opts = append(opts, content.FeedIDs(ids))
	}

	var articleGenerator func() ([]content.Article, error)
	if req.Search != "" {
		if searchProvider != nil {
			articleGenerator = func() ([]content.Article, error) {
				return searchProvider.Search(req.Search, user, opts...)
			}
		}
	} else {
		var skip bool

		switch req.ViewMode {
		case "all_articles":
		case "adaptive":
		case "unread":
			opts = append(opts, content.UnreadOnly)
		case "marked":
			opts = append(opts, content.FavoriteOnly)
		default:
			skip = true
		}

		if !skip {
			articleGenerator = func() ([]content.Article, error) {
				return service.ArticleRepo().ForUser(user, opts...)
			}
		}
	}

	if articleGenerator != nil {
		articles, err := articleGenerator()
		if err != nil {
			return nil, errors.WithMessage(err, "gettting articles")
		}

		if len(articles) > 0 {
			articles = processor.Articles(processors).Process(articles)

			firstID = articles[0].ID
		}

		headlines := headlinesFromArticles(articles, feedTitle, req.ShowContent, req.ShowExcerpt)
		if req.IncludeHeader {
			header := headlinesHeader{Id: req.FeedId, FirstId: firstID, IsCat: req.IsCat}
			hContent := headlinesHeaderContent{}

			hContent = append(hContent, header)
			hContent = append(hContent, headlines)

			return hContent, nil
		}

		return headlines, nil
	}

	return nil, nil
}

func updateArticle(req request, user content.User, service repo.Service) (interface{}, error) {
	if req.Field != 0 && req.Field != 2 {
		return nil, errors.Errorf("Unknown field %d", req.Field)
	}

	articles, err := service.ArticleRepo().ForUser(user,
		content.IDs(req.ArticleIds),
		content.Filters(content.GetUserFilters(user)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "getting usr articles")
	}

	var read, unread, favor, unfavor []content.ArticleID

	for _, a := range articles {
		switch req.Field {
		case 0:
			switch req.Mode {
			case 0:
				if a.Favorite {
					unfavor = append(unfavor, a.ID)
				}
			case 1:
				if !a.Favorite {
					favor = append(favor, a.ID)
				}
			case 2:
				if a.Favorite {
					unfavor = append(unfavor, a.ID)
				} else {
					favor = append(favor, a.ID)
				}
			}
		case 2:
			switch req.Mode {
			case 0:
				if !a.Read {
					read = append(read, a.ID)
				}
			case 1:
				if a.Read {
					unread = append(unread, a.ID)
				}
			case 2:
				if a.Read {
					unread = append(unread, a.ID)
				} else {
					read = append(read, a.ID)
				}
			}
		}
	}

	var updateCount int
	if len(read) > 0 {
		if err = service.ArticleRepo().Read(true, user,
			content.IDs(read),
			content.Filters(content.GetUserFilters(user)),
		); err != nil {
			return nil, errors.WithMessage(err, "marking articles as read")
		}

		updateCount += len(read)
	}

	if len(unread) > 0 {
		if err = service.ArticleRepo().Read(false, user,
			content.IDs(unread),
			content.Filters(content.GetUserFilters(user)),
		); err != nil {
			return nil, errors.WithMessage(err, "marking articles as unread")
		}

		updateCount += len(unread)
	}

	if len(favor) > 0 {
		if err = service.ArticleRepo().Favor(true, user,
			content.IDs(favor),
			content.Filters(content.GetUserFilters(user)),
		); err != nil {
			return nil, errors.WithMessage(err, "marking articles as favorite")
		}

		updateCount += len(favor)
	}

	if len(unfavor) > 0 {
		if err = service.ArticleRepo().Favor(false, user,
			content.IDs(unfavor),
			content.Filters(content.GetUserFilters(user)),
		); err != nil {
			return nil, errors.WithMessage(err, "marking articles as not favorite")
		}

		updateCount += len(unfavor)
	}

	return genericContent{Status: "OK", Updated: int64(updateCount)}, nil
}

func getArticle(
	req request,
	user content.User,
	service repo.Service,
	processors []processor.Article,
) (interface{}, error) {
	articles, err := service.ArticleRepo().ForUser(user,
		content.IDs(req.ArticleIds),
		content.Filters(content.GetUserFilters(user)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "getting user articles")
	}

	feedTitles := map[content.FeedID]string{}

	for _, a := range articles {
		if _, ok := feedTitles[a.FeedID]; !ok {
			f, err := service.FeedRepo().Get(a.FeedID, content.User{})
			if err != nil {
				return nil, errors.Wrapf(err, "getting feed by id %d", a.FeedID)
			}

			feedTitles[a.FeedID] = f.Title
		}
	}

	cContent := articlesContent{}

	for _, a := range articles {
		title := feedTitles[a.FeedID]
		h := article{
			Id:        strconv.FormatInt(int64(a.ID), 10),
			Unread:    !a.Read,
			Marked:    a.Favorite,
			Updated:   a.Date.Unix(),
			Title:     a.Title,
			Link:      a.Link,
			FeedId:    strconv.FormatInt(int64(a.FeedID), 10),
			FeedTitle: title,
			Content:   a.Description,
		}

		cContent = append(cContent, h)
	}

	return cContent, nil
}

func headlinesFromArticles(articles []content.Article, feedTitle string, content, excerpt bool) headlinesContent {
	c := headlinesContent{}
	for _, a := range articles {
		title := feedTitle
		h := headline{
			Id:        a.ID,
			Unread:    !a.Read,
			Marked:    a.Favorite,
			Updated:   a.Date.Unix(),
			IsUpdated: !a.Read,
			Title:     a.Title,
			Link:      a.Link,
			FeedId:    strconv.FormatInt(int64(a.FeedID), 10),
			FeedTitle: title,
		}

		if content {
			h.Content = a.Description
		}

		if excerpt {
			excerpt := search.StripTags(a.Description)
			if len(excerpt) > 100 {
				excerpt = excerpt[:100]
			}

			h.Excerpt = excerpt
		}

		c = append(c, h)
	}

	return c
}
