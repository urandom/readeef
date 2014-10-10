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

		resp := map[string]interface{}{"api_version": 1}

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
