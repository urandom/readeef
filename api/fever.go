package api

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/urandom/readeef"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/util"
)

type Fever struct {
	webfw.BasePatternController
	fm *readeef.FeedManager
}

type feverFeed struct {
	Id         int64  `json:"id"`
	Title      string `json:"title"`
	Url        string `json:"url"`
	SiteUrl    string `json:"site_url"`
	IsSpark    int    `json:"is_spark"`
	UpdateTime int64  `json:"last_updated_on_time"`
}

type feverGroup struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
}

type feverFeedsGroup struct {
	GroupId int64  `json:"group_id"`
	FeedIds string `json:"feed_ids"`
}

type feverItem struct {
	Id            int64  `json:"id"`
	FeedId        int64  `json:"feed_id"`
	Title         string `json:"title"`
	Author        string `json:"author"`
	Html          string `json:"html"`
	Url           string `json:"url"`
	IsSaved       int    `json:"is_saved"`
	IsRead        int    `json:"is_read"`
	CreatedOnTime int64  `json:"created_on_time"`
}

type feverLink struct {
	Id          int64   `json:"id"`
	FeedId      int64   `json:"feed_id"`
	ItemId      int64   `json:"item_id"`
	Temperature float64 `json:"temperature"`
	IsItem      int     `json:"is_item"`
	IsLocal     int     `json:"is_local"`
	IsSaved     int     `json:"is_saved"`
	Title       string  `json:"title"`
	Url         string  `json:"url"`
	ItemIds     string  `json:"item_ids"`
}

func NewFever(fm *readeef.FeedManager) Fever {
	return Fever{
		webfw.NewBasePatternController("/v:version/fever/", webfw.MethodPost, ""),
		fm,
	}
}

var (
	counter  int64 = 0
	tagIdMap       = map[string]int64{}
	idTagMap       = map[int64]string{}
)

func (con Fever) Handler(c context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		var user readeef.User

		db := readeef.GetDB(c)

		err = r.ParseForm()

		if err == nil {
			user = getUser(db, r.FormValue("api_key"), webfw.GetLogger(c))
		}

		resp := map[string]interface{}{"api_version": 2}

		switch {
		default:
			if user.Login == "" {
				resp["auth"] = 0
				break
			}

			now := time.Now().Unix()

			resp["auth"] = 1
			resp["last_refreshed_on_time"] = now

			if _, ok := r.Form["groups"]; ok {
				readeef.Debug.Println("Fetching fever groups")

				groups := []feverGroup{feverGroup{Id: 1, Title: "All"}}

				resp["groups"] = groups

				var feeds []readeef.Feed

				feeds, err = db.GetUserFeeds(user)
				if err != nil {
					break
				}

				resp["feeds_groups"] = getFeedsGroups(feeds)
			}

			if _, ok := r.Form["feeds"]; ok {
				readeef.Debug.Println("Fetching fever feeds")

				var feeds []readeef.Feed
				var feverFeeds []feverFeed

				feeds, err = db.GetUserFeeds(user)

				if err != nil {
					break
				}

				for _, f := range feeds {
					feed := feverFeed{
						Id: f.Id, Title: f.Title, Url: f.Link, SiteUrl: f.SiteLink, UpdateTime: now,
					}

					feverFeeds = append(feverFeeds, feed)
				}

				resp["feeds"] = feverFeeds
				resp["feeds_groups"] = getFeedsGroups(feeds)
			}

			if _, ok := r.Form["unread_item_ids"]; ok {
				readeef.Debug.Println("Fetching unread fever item ids")

				var ids []int64

				ids, err = db.GetAllUnreadUserArticleIds(user)
				if err != nil {
					break
				}

				buf := util.BufferPool.GetBuffer()
				defer util.BufferPool.Put(buf)

				for i, id := range ids {
					if i != 0 {
						buf.WriteString(",")
					}

					buf.WriteString(strconv.FormatInt(id, 10))
				}

				resp["unread_item_ids"] = buf.String()
			}

			if _, ok := r.Form["saved_item_ids"]; ok {
				readeef.Debug.Println("Fetching saved fever item ids")

				var ids []int64

				ids, err = db.GetAllFavoriteUserArticleIds(user)
				if err != nil {
					break
				}

				buf := util.BufferPool.GetBuffer()
				defer util.BufferPool.Put(buf)

				for i, id := range ids {
					if i != 0 {
						buf.WriteString(",")
					}

					buf.WriteString(strconv.FormatInt(id, 10))
				}

				resp["saved_item_ids"] = buf.String()
			}

			if _, ok := r.Form["items"]; ok {
				readeef.Debug.Println("Fetching fever items")

				var count, since, max int64

				count, err = db.GetUserArticleCount(user)
				if err != nil {
					break
				}

				items := []feverItem{}
				if count > 0 {
					if val, ok := r.Form["since_id"]; ok {
						since, err = strconv.ParseInt(val[0], 10, 64)
						if err != nil {
							err = nil
							since = 0
						}
					}

					if val, ok := r.Form["max_id"]; ok {
						max, err = strconv.ParseInt(val[0], 10, 64)
						if err != nil {
							err = nil
							since = 0
						}
					}

					var articles []readeef.Article
					if withIds, ok := r.Form["with_ids"]; ok {
						stringIds := strings.Split(withIds[0], ",")
						ids := make([]int64, len(stringIds))

						for i, stringId := range stringIds {
							stringId = strings.TrimSpace(stringId)

							if id, err := strconv.ParseInt(stringId, 10, 64); err == nil {
								ids[i] = id
							}
						}

						articles, err = db.GetSpecificUserArticles(user, ids...)
					} else if max > 0 {
						articles, err = db.GetUnorderedUserArticlesDesc(user, max, 50, 0)
					} else {
						articles, err = db.GetUnorderedUserArticles(user, since, 50, 0)
					}

					if err != nil {
						break
					}

					for _, a := range articles {
						item := feverItem{
							Id: a.Id, FeedId: a.FeedId, Title: a.Title, Html: a.Description,
							Url: a.Link, CreatedOnTime: a.Date.Unix(),
						}
						if a.Read {
							item.IsRead = 1
						}
						if a.Favorite {
							item.IsSaved = 1
						}
						items = append(items, item)
					}
				}

				resp["total_items"] = count
				resp["items"] = items
			}

			if _, ok := r.Form["links"]; ok {
				readeef.Debug.Println("Fetching fever links")
				offset, _ := strconv.ParseInt(r.FormValue("offset"), 10, 64)

				rng, e := strconv.ParseInt(r.FormValue("range"), 10, 64)
				if e != nil {
					rng = 7
				}

				page := int64(1)
				page, err = strconv.ParseInt(r.FormValue("page"), 10, 64)
				if e != nil {
					break
				}

				if page > 3 {
					resp["links"] = []feverLink{}
					break
				}

				var articles []readeef.Article
				var from, to time.Time

				if offset == 0 {
					from = time.Now().AddDate(0, 0, int(-1*rng))
					to = time.Now()
				} else {
					from = time.Now().AddDate(0, 0, int(-1*rng-offset))
					to = time.Now().AddDate(0, 0, int(-1*offset))
				}

				timeRange := readeef.TimeRange{From: from, To: to}

				articles, err = db.GetScoredUserArticlesDesc(user, timeRange, 50, 50*int(page-1))
				if err != nil {
					break
				}

				links := make([]feverLink, len(articles))
				for i, a := range articles {
					link := feverLink{
						Id: a.Id, FeedId: a.FeedId, ItemId: a.Id, Temperature: math.Log10(float64(a.Score)) / math.Log10(1.1),
						IsItem: 1, IsLocal: 1, Title: a.Title, Url: a.Link, ItemIds: fmt.Sprintf("%d", a.Id),
					}

					if a.Favorite {
						link.IsSaved = 1
					}

					links[i] = link
				}
				resp["links"] = links
			}

			if val := r.PostFormValue("unread_recently_read"); val == "1" {
				readeef.Debug.Println("Marking recently read fever items as unread")

				t := time.Now().Add(-24 * time.Hour)
				err = db.MarkNewerUserArticlesByDateAsUnread(user, t, true)
				if err != nil {
					break
				}
			}

			if val := r.PostFormValue("mark"); val != "" {
				if val == "item" {
					readeef.Debug.Printf("Marking fever item '%s' as '%s'\n", r.PostFormValue("id"), r.PostFormValue("as"))

					var id int64
					var article readeef.Article

					id, err = strconv.ParseInt(r.PostFormValue("id"), 10, 64)
					if err != nil {
						break
					}

					article, err = db.GetFeedArticle(id, user)
					if err != nil {
						break
					}

					switch r.PostFormValue("as") {
					case "read":
						err = db.MarkUserArticlesAsRead(user, []readeef.Article{article}, true)
					case "saved":
						err = db.MarkUserArticlesAsFavorite(user, []readeef.Article{article}, true)
					case "unsaved":
						err = db.MarkUserArticlesAsFavorite(user, []readeef.Article{article}, false)
					default:
						err = errors.New("Unknown 'as' action")
					}
				} else if val == "feed" || val == "group" {
					readeef.Debug.Printf("Marking fever %s '%s' as '%s'\n", val, r.PostFormValue("id"), r.PostFormValue("as"))
					if r.PostFormValue("as") != "read" {
						err = errors.New("Unknown 'as' action")
						break
					}

					var id, timestamp int64

					id, err = strconv.ParseInt(r.PostFormValue("id"), 10, 64)
					if err != nil {
						break
					}

					timestamp, err = strconv.ParseInt(r.PostFormValue("before"), 10, 64)
					if err != nil {
						break
					}

					t := time.Unix(timestamp, 0)

					if val == "feed" {
						var feed readeef.Feed

						feed, err = db.GetUserFeed(id, user)
						if err != nil {
							break
						}

						err = db.MarkFeedArticlesByDateAsRead(feed, t, true)
					} else if val == "group" {
						if id == 1 || id == 0 {
							err = db.MarkUserArticlesByDateAsRead(user, t, true)
						} else {
							err = errors.New(fmt.Sprintf("Unknown group %d\n", id))
						}
					}
				}
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

	})
}

func getUser(db readeef.DB, md5hex string, log webfw.Logger) readeef.User {
	md5, err := hex.DecodeString(md5hex)

	if err != nil {
		log.Printf("Error decoding hex api_key")
		return readeef.User{}
	}

	user, err := db.GetUserByMD5Api(md5)
	if err != nil {
		log.Printf("Error getting user by md5api field: %v\n", err)
	}
	return user
}

func getFeedsGroups(feeds []readeef.Feed) []feverFeedsGroup {
	buf := util.BufferPool.GetBuffer()
	defer util.BufferPool.Put(buf)

	for i, f := range feeds {
		if i != 0 {
			buf.WriteString(",")
		}

		buf.WriteString(strconv.FormatInt(f.Id, 10))
	}

	return []feverFeedsGroup{feverFeedsGroup{GroupId: 1, FeedIds: buf.String()}}
}
