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

		if user.Login == "" {
			resp["auth"] = 0
		} else {
			now := time.Now().Unix()

			resp["auth"] = 1
			resp["last_refreshed_on_time"] = now

			if _, ok := r.Form["groups"]; ok {
				readeef.Debug.Println("Fetching fever groups")

				resp["groups"] = []feverGroup{
					feverGroup{Id: 1, Title: "All"},
				}

				var feeds []readeef.Feed

				feeds, err = db.GetUserFeeds(user)
				if err == nil {
					ids := []string{}
					for _, f := range feeds {
						ids = append(ids, strconv.FormatInt(f.Id, 10))
					}

					resp["feeds_groups"] = []feverFeedsGroup{
						feverFeedsGroup{GroupId: 1, FeedIds: strings.Join(ids, ",")},
					}
				}
			}

			if _, ok := r.Form["feeds"]; ok {
				readeef.Debug.Println("Fetching fever feeds")

				var feeds []readeef.Feed
				var feverFeeds []feverFeed

				feeds, err = db.GetUserFeeds(user)

				if err == nil {
					for _, f := range feeds {
						feed := feverFeed{
							Id: f.Id, Title: f.Title, Url: f.Link, SiteUrl: f.SiteLink, UpdateTime: now,
						}

						feverFeeds = append(feverFeeds, feed)
					}

					resp["feeds"] = feverFeeds

					ids := []string{}
					for _, f := range feeds {
						ids = append(ids, strconv.FormatInt(f.Id, 10))
					}

					resp["feeds_groups"] = []feverFeedsGroup{
						feverFeedsGroup{GroupId: 1, FeedIds: strings.Join(ids, ",")},
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
