package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base/search"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/util"
)

const (
	TTRSS_API_STATUS_OK  = 0
	TTRSS_API_STATUS_ERR = 1
	TTRSS_VERSION        = "1.8.0"
	TTRSS_API_LEVEL      = 12
	TTRSS_SESSION_ID     = "TTRSS_SESSION_ID"
	TTRSS_USER_NAME      = "TTRSS_USER_NAME"

	TTRSS_FAVORITE_ID = -1
	TTRSS_ALL_ID      = -4
)

type TtRss struct {
	webfw.BasePatternController
	sp content.SearchProvider
	ap []ArticleProcessor
}

type ttRssRequest struct {
	Op            string         `json:"op"`
	Sid           string         `json:"sid"`
	User          string         `json:"user"`
	Password      string         `json:"password"`
	OutputMode    string         `json:"output_mode"`
	UnreadOnly    bool           `json:"unread_only"`
	Limit         int            `json:"limit"`
	Offset        int            `json:"offset"`
	CatId         int            `json:"cat_id"`
	FeedId        data.FeedId    `json:"feed_id"`
	Skip          int            `json:"skip"`
	IsCat         bool           `json:"is_cat"`
	ShowContent   bool           `json:"show_content"`
	ShowExcerpt   bool           `json:"show_excerpt"`
	ViewMode      string         `json:"view_mode"`
	SinceId       data.ArticleId `json:"since_id"`
	Sanitize      bool           `json:"sanitize"`
	HasSandbox    bool           `json:"has_sandbox"`
	IncludeHeader bool           `json:"include_header"`
	OrderBy       string         `json:"order_by"`
	Search        string         `json:"search"`
}

type ttRssResponse struct {
	Seq     int64           `json:"seq"`
	Status  int             `json:"status"`
	Content json.RawMessage `json:"content"`
}

type ttRssGenericContent struct {
	Error     string      `json:"error,omitempty"`
	Level     int         `json:"level,omitempty"`
	ApiLevel  int         `json:"api_level,omitempty"`
	Version   string      `json:"version,omitempty"`
	SessionId string      `json:"session_id,omitempty"`
	Status    interface{} `json:"status,omitempty"`
	Unread    int64       `json:"unread,omitempty"`
}

type ttRssCountersContent []ttRssCounter

type ttRssCounter struct {
	Id         string `json:"id"`
	Counter    int64  `json:"counter"`
	AuxCounter int64  `json:"auxcounter,omitempty"`
}

type ttRssFeedsContent []ttRssFeed

type ttRssFeed struct {
	Id          data.FeedId `json:"id"`
	Title       string      `json:"title"`
	Unread      int64       `json:"unread"`
	CatId       int         `json:"cat_id"`
	FeedUrl     string      `json:"feed_url,omitempty"`
	LastUpdated int64       `json:"last_updated,omitempty"`
	OrderId     int         `json:"order_id,omitempty"`
}

type ttRssHeadlinesHeaderContent []interface{}
type ttRssHeadlinesContent []ttRssHeadline

type ttRssHeadline struct {
	Id        data.ArticleId `json:"id"`
	Unread    bool           `json:"unread"`
	Marked    bool           `json:"marked"`
	Updated   int64          `json:"updated"`
	IsUpdated bool           `json:"is_updated"`
	Title     string         `json:"title"`
	Link      string         `json:"link"`
	FeedId    data.FeedId    `json:"feed_id"`
	Excerpt   string         `json:"excerpt"`
	Content   string         `json:"content"`
	FeedTitle string         `json:"feed_title"`

	Tags   []string `json:"tags"`
	Labels []string `json:"labels"`
}

type ttRssHeadlinesHeader struct {
	Id      data.FeedId    `json:"id"`
	FirstId data.ArticleId `json:"first_id"`
	IsCat   bool           `json:"is_cat"`
}

func NewTtRss(sp content.SearchProvider, ap []ArticleProcessor) TtRss {
	return TtRss{
		webfw.NewBasePatternController("/v"+strconv.Itoa(TTRSS_API_LEVEL)+"/tt-rss/", webfw.MethodPost, ""),
		sp, ap,
	}
}

func (controller TtRss) Handler(c context.Context) http.Handler {
	repo := readeef.GetRepo(c)
	logger := webfw.GetLogger(c)
	config := readeef.GetConfig(c)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		req := ttRssRequest{}
		dec := json.NewDecoder(r.Body)
		sess := webfw.GetSession(c, r)

		resp := ttRssResponse{}

		var err error
		var errType string
		var user content.User
		var con interface{}

		switch {
		default:
			err = r.ParseForm()
			if err != nil {
				break
			}

			seq, _ := strconv.ParseInt(r.Form.Get("seq"), 10, 64)

			resp.Seq = seq

			err = dec.Decode(&req)
			if err != nil {
				err = fmt.Errorf("Error decoding JSON request: %v", err)
				break
			}

			if req.Op != "login" && req.Op != "isLoggedIn" {
				if id, ok := sess.Get(TTRSS_SESSION_ID); ok && id != "" && id == req.Sid {
					userName, _ := sess.Get(TTRSS_USER_NAME)
					if login, ok := userName.(string); ok {
						user = repo.UserByLogin(data.Login(login))
						if repo.Err() != nil {
							errType = "NOT_LOGGED_IN"
						}
					} else {
						errType = "NOT_LOGGED_IN"
					}
				} else {
					errType = "NOT_LOGGED_IN"
				}
			}

			if errType != "" {
				break
			}

			switch req.Op {
			case "getApiLevel":
				con = ttRssGenericContent{Level: TTRSS_API_LEVEL}
			case "getVersion":
				con = ttRssGenericContent{Version: TTRSS_VERSION}
			case "login":
				user = repo.UserByLogin(data.Login(req.User))
				if repo.Err() != nil {
					errType = "LOGIN_ERROR"
					break
				}

				if !user.Authenticate(req.Password, []byte(config.Auth.Secret)) {
					errType = "LOGIN_ERROR"
					break
				}

				sessId := util.UUID()
				sess.Set(TTRSS_SESSION_ID, sessId)
				sess.Set(TTRSS_USER_NAME, req.User)

				con = ttRssGenericContent{
					ApiLevel:  TTRSS_API_LEVEL,
					SessionId: sessId,
				}
			case "logout":
				sess.Delete(TTRSS_SESSION_ID)
				sess.Delete(TTRSS_USER_NAME)

				con = ttRssGenericContent{Status: "OK"}
			case "isLoggedIn":
				if id, ok := sess.Get(TTRSS_SESSION_ID); ok && id != "" {
					con = ttRssGenericContent{Status: true}
				} else {
					con = ttRssGenericContent{Status: false}
				}
			case "getUnread":
				var count int64
				counted := false

				if fid := r.Form.Get("feed_id"); fid != "" {
					// Can't handle categories, they are integer ids
					if isTag := r.Form.Get("is_cat"); isTag == "" {
						if feedId, err := strconv.ParseInt(fid, 10, 64); err == nil {
							feed := user.FeedById(data.FeedId(feedId))
							count = feed.Count(data.ArticleCountOptions{UnreadOnly: true})
							if feed.HasErr() {
								err = feed.Err()
								break
							}

							counted = true
						}
					}
				}

				if !counted {
					count = user.Count(data.ArticleCountOptions{UnreadOnly: true})
					if user.HasErr() {
						err = fmt.Errorf("Error getting all unread article ids: %v\n", user.Err())
					}
				}

				con = ttRssGenericContent{Unread: count}
			case "getCounters":
				if req.OutputMode == "" {
					req.OutputMode = "flc"
				}
				cContent := ttRssCountersContent{}

				o := data.ArticleCountOptions{UnreadOnly: true}
				unreadCount := user.Count(o)
				cContent = append(cContent,
					ttRssCounter{Id: "global-unread", Counter: unreadCount})

				feeds := user.AllFeeds()
				cContent = append(cContent,
					ttRssCounter{Id: "subscribed-feeds", Counter: int64(len(feeds))})

				cContent = append(cContent,
					ttRssCounter{Id: strconv.Itoa(TTRSS_FAVORITE_ID),
						Counter:    0,
						AuxCounter: int64(len(user.AllFavoriteArticleIds()))})

				cContent = append(cContent,
					ttRssCounter{Id: strconv.Itoa(TTRSS_ALL_ID),
						Counter:    user.Count(),
						AuxCounter: unreadCount})

				if strings.Contains(req.OutputMode, "f") {
					for _, f := range feeds {
						cContent = append(cContent,
							ttRssCounter{Id: strconv.FormatInt(int64(f.Data().Id), 10), Counter: f.Count(o)},
						)

					}
				}

				if user.HasErr() {
					err = fmt.Errorf("Error getting user counters: %v\n", user.Err())
				}

				con = cContent
			case "getFeeds":
				fContent := ttRssFeedsContent{}

				if req.CatId == TTRSS_FAVORITE_ID || req.CatId == TTRSS_ALL_ID {
					if !req.UnreadOnly {
						fContent = append(fContent, ttRssFeed{
							Id:     TTRSS_FAVORITE_ID,
							Title:  "Starred articles",
							Unread: 0,
							CatId:  TTRSS_FAVORITE_ID,
						})
					}

					unread := user.Count(data.ArticleCountOptions{UnreadOnly: true})

					if unread > 0 || !req.UnreadOnly {
						fContent = append(fContent, ttRssFeed{
							Id:     TTRSS_ALL_ID,
							Title:  "All articles",
							Unread: unread,
							CatId:  TTRSS_FAVORITE_ID,
						})
					}
				}

				if req.CatId == 0 || req.CatId == TTRSS_ALL_ID {
					feeds := user.AllFeeds()
					o := data.ArticleCountOptions{UnreadOnly: true}
					for i := range feeds {
						if req.Limit > 0 {
							if i < req.Offset || i >= req.Limit+req.Offset {
								continue
							}
						}

						d := feeds[i].Data()
						unread := feeds[i].Count(o)

						if unread > 0 || !req.UnreadOnly {
							fContent = append(fContent, ttRssFeed{
								Id:          d.Id,
								Title:       d.Title,
								FeedUrl:     d.Link,
								CatId:       0,
								Unread:      unread,
								LastUpdated: time.Now().Unix(),
								OrderId:     0,
							})
						}
					}
				}

				if user.HasErr() {
					err = fmt.Errorf("Error getting user feeds: %v\n", user.Err())
				}

				con = fContent
			case "getCategories":
				con = ttRssFeedsContent{}
			case "getHeadlines":
				if req.FeedId == 0 {
					errType = "INCORRECT_USAGE"
					break
				}

				limit := req.Limit
				if limit == 0 {
					limit = 200
				}

				var articles []content.UserArticle
				var articleRepo content.ArticleRepo
				var feedTitle string
				firstId := data.ArticleId(0)
				o := data.ArticleQueryOptions{Limit: limit, Offset: req.Skip}

				if req.FeedId == TTRSS_FAVORITE_ID {
					ttRssSetupSorting(req, user)
					o.FavoriteOnly = true
					articleRepo = user
					feedTitle = "Starred articles"
				} else if req.FeedId == TTRSS_ALL_ID {
					ttRssSetupSorting(req, user)
					articleRepo = user
					feedTitle = "All articles"
				} else if req.FeedId > 0 {
					feed := user.FeedById(req.FeedId)

					ttRssSetupSorting(req, feed)
					articleRepo = feed
					feedTitle = feed.Data().Title
				}

				if articleRepo != nil {
					if req.Search != "" {
						if controller.sp != nil {
							if as, ok := articleRepo.(content.ArticleSearch); ok {
								articles = as.Query(req.Search, controller.sp, limit, req.Skip)
							}
						}
					} else {
						var skip bool

						switch req.ViewMode {
						case "all_articles":
						case "unread":
							o.UnreadOnly = true
						default:
							skip = true
						}

						if !skip {
							articles = articleRepo.Articles(o)
						}
					}
				}

				if len(articles) > 0 {
					firstId = articles[0].Data().Id
				}

				if req.IncludeHeader {
					header := ttRssHeadlinesHeader{Id: req.FeedId, FirstId: firstId}
					hContent := ttRssHeadlinesHeaderContent{}

					hContent = append(hContent, header)
					hContent = append(hContent, ttRssHeadlinesFromArticles(req, articles, feedTitle))

					con = hContent
				} else {
					con = ttRssHeadlinesFromArticles(req, articles, feedTitle)
				}
			}
		}

		if err == nil && errType == "" {
			resp.Status = TTRSS_API_STATUS_OK
		} else {
			resp.Status = TTRSS_API_STATUS_ERR
			switch v := con.(type) {
			case ttRssGenericContent:
				v.Error = errType
			}
		}

		var b []byte
		b, err = json.Marshal(con)
		if err == nil {
			resp.Content = json.RawMessage(b)
		}

		b, err = json.Marshal(&resp)

		if err == nil {
			w.Header().Set("Content-Type", "text/json")
			w.Write(b)
		} else {
			logger.Print(fmt.Errorf("TT-RSS error %s: %v", req.Op, err))

			w.WriteHeader(http.StatusInternalServerError)
		}

	})
}

func ttRssSetupSorting(req ttRssRequest, sorting content.ArticleSorting) {
	switch req.OrderBy {
	case "date_reverse":
		sorting.SortingByDate()
		sorting.Order(data.AscendingOrder)
	default:
		sorting.SortingByDate()
		sorting.Order(data.DescendingOrder)
	}
}

func ttRssHeadlinesFromArticles(req ttRssRequest, articles []content.UserArticle, feedTitle string) (c ttRssHeadlinesContent) {
	for _, a := range articles {
		d := a.Data()
		h := ttRssHeadline{
			Id:        d.Id,
			Unread:    !d.Read,
			Marked:    d.Favorite,
			IsUpdated: !d.Read,
			Title:     d.Title,
			Link:      d.Link,
			FeedId:    d.FeedId,
			FeedTitle: feedTitle,
		}

		if req.ShowExcerpt {
			excerpt := search.StripTags(d.Description)
			if len(excerpt) > 100 {
				excerpt = excerpt[:100]
			}

			h.Excerpt = excerpt
		}

		if req.ShowContent {
			h.Content = d.Description
		}

		c = append(c, h)
	}
	return
}
