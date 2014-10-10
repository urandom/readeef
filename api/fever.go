package api

import (
	"encoding/hex"
	"encoding/json"
	"log"
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
	webfw.BaseController
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

func NewFever(fm *readeef.FeedManager) Fever {
	return Fever{
		webfw.NewBaseController("/v:version/fever/", webfw.MethodPost, ""),
		fm,
	}
}

var (
	counter  int64 = 0
	tagIdMap       = map[string]int64{}
	idTagMap       = map[int64]string{}
)

func (con Fever) Handler(c context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

				var tags []string

				tags, err = db.GetUserTags(user)

				if err != nil {
					break
				}

				groups := []feverGroup{}
				for _, tag := range tags {
					groups = append(groups, tagToGroup(tag))
				}

				resp["groups"] = groups

				var feeds []readeef.Feed

				feeds, err = db.GetUserTagsFeeds(user)
				if err != nil {
					break
				}

				resp["feeds_groups"] = getFeedsGroups(feeds)
			}

			if _, ok := r.Form["feeds"]; ok {
				readeef.Debug.Println("Fetching fever feeds")

				var feeds []readeef.Feed
				var feverFeeds []feverFeed

				feeds, err = db.GetUserTagsFeeds(user)

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

			if _, ok := r.Form["items"]; ok {
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

	}
}

func getUser(db readeef.DB, md5hex string, log *log.Logger) readeef.User {
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

func tagToGroup(tag string) feverGroup {
	if id, ok := tagIdMap[tag]; ok {
		return feverGroup{Id: id, Title: tag}
	} else {
		counter++
		idTagMap[counter] = tag
		tagIdMap[tag] = counter

		return feverGroup{Id: id, Title: tag}
	}
}

func getFeedsGroups(feeds []readeef.Feed) []feverFeedsGroup {
	feedGroupMap := map[int64][]string{}

	for _, f := range feeds {
		for _, tag := range f.Tags {
			id := tagIdMap[tag]
			feedGroupMap[id] = append(feedGroupMap[id], strconv.FormatInt(f.Id, 10))
		}
	}

	feedsGroups := []feverFeedsGroup{}
	for group, feeds := range feedGroupMap {
		feedsGroups = append(feedsGroups,
			feverFeedsGroup{GroupId: group, FeedIds: strings.Join(feeds, ",")})
	}

	return feedsGroups
}
