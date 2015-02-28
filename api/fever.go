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
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/info"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/util"
)

type Fever struct {
	webfw.BasePatternController
	fm *readeef.FeedManager
}

type feverFeed struct {
	Id         info.FeedId `json:"id"`
	Title      string      `json:"title"`
	Url        string      `json:"url"`
	SiteUrl    string      `json:"site_url"`
	IsSpark    int         `json:"is_spark"`
	UpdateTime int64       `json:"last_updated_on_time"`
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
	Id            info.ArticleId `json:"id"`
	FeedId        info.FeedId    `json:"feed_id"`
	Title         string         `json:"title"`
	Author        string         `json:"author"`
	Html          string         `json:"html"`
	Url           string         `json:"url"`
	IsSaved       int            `json:"is_saved"`
	IsRead        int            `json:"is_read"`
	CreatedOnTime int64          `json:"created_on_time"`
}

type feverLink struct {
	Id          info.ArticleId `json:"id"`
	FeedId      info.FeedId    `json:"feed_id"`
	ItemId      info.ArticleId `json:"item_id"`
	Temperature float64        `json:"temperature"`
	IsItem      int            `json:"is_item"`
	IsLocal     int            `json:"is_local"`
	IsSaved     int            `json:"is_saved"`
	Title       string         `json:"title"`
	Url         string         `json:"url"`
	ItemIds     string         `json:"item_ids"`
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
	repo := readeef.GetRepo(c)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := webfw.GetLogger(c)

		var err error
		var user content.User

		err = r.ParseForm()

		if err == nil {
			user = getReadeefUser(repo, r.FormValue("api_key"), webfw.GetLogger(c))
		}

		resp := map[string]interface{}{"api_version": 2}

		switch {
		default:
			if user == nil || user.Validate() != nil {
				resp["auth"] = 0
				break
			}

			now := time.Now().Unix()

			resp["auth"] = 1
			resp["last_refreshed_on_time"] = now

			if _, ok := r.Form["groups"]; ok {
				logger.Infoln("Fetching fever groups")

				groups := []feverGroup{feverGroup{Id: 1, Title: "All"}}

				resp["groups"] = groups

				feeds := user.AllFeeds()
				err = user.Err()

				if err != nil {
					break
				}

				resp["feeds_groups"] = getFeedsGroups(feeds)
			}

			if _, ok := r.Form["feeds"]; ok {
				logger.Infoln("Fetching fever feeds")

				var feverFeeds []feverFeed

				feeds := user.AllFeeds()
				err = user.Err()

				if err != nil {
					break
				}

				for i := range feeds {
					in := feeds[i].Info()
					feed := feverFeed{
						Id: in.Id, Title: in.Title, Url: in.Link, SiteUrl: in.SiteLink, UpdateTime: now,
					}

					feverFeeds = append(feverFeeds, feed)
				}

				resp["feeds"] = feverFeeds
				resp["feeds_groups"] = getFeedsGroups(feeds)
			}

			if _, ok := r.Form["unread_item_ids"]; ok {
				logger.Infoln("Fetching unread fever item ids")

				ids := user.AllUnreadArticleIds()
				err = user.Err()
				if err != nil {
					break
				}

				buf := util.BufferPool.GetBuffer()
				defer util.BufferPool.Put(buf)

				for i := range ids {
					if i != 0 {
						buf.WriteString(",")
					}

					buf.WriteString(strconv.FormatInt(int64(ids[i]), 10))
				}

				resp["unread_item_ids"] = buf.String()
			}

			if _, ok := r.Form["saved_item_ids"]; ok {
				logger.Infoln("Fetching saved fever item ids")

				ids := user.AllFavoriteIds()
				err = user.Err()
				if err != nil {
					break
				}

				buf := util.BufferPool.GetBuffer()
				defer util.BufferPool.Put(buf)

				for i := range ids {
					if i != 0 {
						buf.WriteString(",")
					}

					buf.WriteString(strconv.FormatInt(int64(ids[i]), 10))
				}

				resp["saved_item_ids"] = buf.String()
			}

			if _, ok := r.Form["items"]; ok {
				logger.Infoln("Fetching fever items")

				var count, since, max int64

				count, err = user.ArticleCount(), user.Err()
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

					var articles []content.UserArticle
					if withIds, ok := r.Form["with_ids"]; ok {
						stringIds := strings.Split(withIds[0], ",")
						ids := make([]info.ArticleId, 0, len(stringIds))

						for _, stringId := range stringIds {
							stringId = strings.TrimSpace(stringId)

							if id, err := strconv.ParseInt(stringId, 10, 64); err == nil {
								ids = append(ids, info.ArticleId(id))
							}
						}

						articles, err = user.ArticlesById(ids), user.Err()
					} else if max > 0 {
						user.Order(info.DescendingOrder)
						articles, err = user.ArticlesOrderedById(info.ArticleId(max), 50, 0), user.Err()
					} else {
						user.Order(info.AscendingOrder)
						articles, err = user.ArticlesOrderedById(info.ArticleId(since), 50, 0), user.Err()
					}

					if err != nil {
						break
					}

					for i := range articles {
						in := articles[i].Info()
						item := feverItem{
							Id: in.Id, FeedId: in.FeedId, Title: in.Title, Html: in.Description,
							Url: in.Link, CreatedOnTime: in.Date.Unix(),
						}
						if in.Read {
							item.IsRead = 1
						}
						if in.Favorite {
							item.IsSaved = 1
						}
						items = append(items, item)
					}
				}

				resp["total_items"] = count
				resp["items"] = items
			}

			if _, ok := r.Form["links"]; ok {
				logger.Infoln("Fetching fever links")
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

				var articles []content.ScoredArticle
				var from, to time.Time

				if offset == 0 {
					from = time.Now().AddDate(0, 0, int(-1*rng))
					to = time.Now()
				} else {
					from = time.Now().AddDate(0, 0, int(-1*rng-offset))
					to = time.Now().AddDate(0, 0, int(-1*offset))
				}

				articles, err = user.ScoredArticles(from, to, 50, 50*int(page-1)), user.Err()
				if err != nil {
					break
				}

				links := make([]feverLink, len(articles))
				for i := range articles {
					in := articles[i].Info()
					link := feverLink{
						Id: in.Id, FeedId: in.FeedId, ItemId: in.Id, Temperature: math.Log10(float64(in.Score)) / math.Log10(1.1),
						IsItem: 1, IsLocal: 1, Title: in.Title, Url: in.Link, ItemIds: fmt.Sprintf("%d", in.Id),
					}

					if in.Favorite {
						link.IsSaved = 1
					}

					links[i] = link
				}
				resp["links"] = links
			}

			if val := r.PostFormValue("unread_recently_read"); val == "1" {
				logger.Infoln("Marking recently read fever items as unread")

				t := time.Now().Add(-24 * time.Hour)
				user.ReadAfter(t, true)
				err = user.Err()
				if err != nil {
					break
				}
			}

			if val := r.PostFormValue("mark"); val != "" {
				if val == "item" {
					logger.Infof("Marking fever item '%s' as '%s'\n", r.PostFormValue("id"), r.PostFormValue("as"))

					var id int64
					var article content.UserArticle

					id, err = strconv.ParseInt(r.PostFormValue("id"), 10, 64)
					if err != nil {
						break
					}

					article, err = user.ArticleById(info.ArticleId(id)), user.Err()
					if err != nil {
						break
					}

					switch r.PostFormValue("as") {
					case "read":
						article.Read(true)
					case "saved":
						article.Favorite(true)
					case "unsaved":
						article.Favorite(false)
					default:
						err = errors.New("Unknown 'as' action")
					}
					if err == nil {
						err = article.Err()
					}
				} else if val == "feed" || val == "group" {
					logger.Infof("Marking fever %s '%s' as '%s'\n", val, r.PostFormValue("id"), r.PostFormValue("as"))
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
						var feed content.UserFeed

						feed, err = user.FeedById(info.FeedId(id)), user.Err()
						if err != nil {
							break
						}

						feed.ReadBefore(t, true)
						err = feed.Err()
					} else if val == "group" {
						if id == 1 || id == 0 {
							user.ReadBefore(t, true)
							err = user.Err()
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

func getReadeefUser(repo content.Repo, md5hex string, log webfw.Logger) content.User {
	md5, err := hex.DecodeString(md5hex)

	if err != nil {
		log.Printf("Error decoding hex api_key")
		return nil
	}

	user := repo.UserByMD5Api(md5)
	if repo.HasErr() {
		log.Printf("Error getting user by md5api field: %v\n", repo.Err())
	}
	return user
}

func getFeedsGroups(feeds []content.UserFeed) []feverFeedsGroup {
	buf := util.BufferPool.GetBuffer()
	defer util.BufferPool.Put(buf)

	for i := range feeds {
		if i != 0 {
			buf.WriteString(",")
		}

		buf.WriteString(strconv.FormatInt(int64(feeds[i].Info().Id), 10))
	}

	return []feverFeedsGroup{feverFeedsGroup{GroupId: 1, FeedIds: buf.String()}}
}
