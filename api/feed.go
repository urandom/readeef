package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/parser"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/util"
)

type Feed struct {
	fm *readeef.FeedManager
	sp content.SearchProvider
}

func NewFeed(fm *readeef.FeedManager, sp content.SearchProvider) Feed {
	return Feed{fm: fm, sp: sp}
}

type listFeedsProcessor struct {
	user content.User
}

type discoverFeedsProcessor struct {
	Link string `json:"link"`

	fm   *readeef.FeedManager
	user content.User
}

type parseOpmlProcessor struct {
	Opml string `json:"opml"`

	fm   *readeef.FeedManager
	user content.User
}

type addFeedsProcessor struct {
	Links []string `json:"links"`

	fm   *readeef.FeedManager
	user content.User
}

type removeFeedProcessor struct {
	Id data.FeedId `json:"id"`

	fm   *readeef.FeedManager
	user content.User
}

type getFeedTagsProcessor struct {
	Id data.FeedId `json:"id"`

	user content.User
}

type setFeedTagsProcessor struct {
	Id   data.FeedId     `json:"id"`
	Tags []data.TagValue `json:"tags"`

	user content.User
}

type markFeedAsReadProcessor struct {
	Id        string `json:"id"`
	Timestamp int64  `json:"timestamp"`

	user content.User
}

type getFeedArticlesProcessor struct {
	Id         string `json:"id"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
	NewerFirst bool   `json:"newerFirst"`
	UnreadOnly bool   `json:"unreadOnly"`

	user content.User
	sp   content.SearchProvider
}

var (
	errTypeNoAbsolute = "error-no-absolute"
	errTypeNoFeed     = "error-no-feed"
)

func (con Feed) Patterns() []webfw.MethodIdentifierTuple {
	prefix := "/v:version/feed/"

	return []webfw.MethodIdentifierTuple{
		webfw.MethodIdentifierTuple{prefix, webfw.MethodGet, "list"},
		webfw.MethodIdentifierTuple{prefix, webfw.MethodPost, "add"},
		webfw.MethodIdentifierTuple{prefix + ":feed-id", webfw.MethodDelete, "remove"},
		webfw.MethodIdentifierTuple{prefix + ":feed-id/tags", webfw.MethodGet | webfw.MethodPost, "tags"},
		webfw.MethodIdentifierTuple{prefix + ":feed-id/read/:timestamp", webfw.MethodPost, "read"},
		webfw.MethodIdentifierTuple{prefix + ":feed-id/articles/:limit/:offset/:newer-first/:unread-only", webfw.MethodGet, "articles"},

		webfw.MethodIdentifierTuple{prefix + "discover", webfw.MethodGet, "discover"},
		webfw.MethodIdentifierTuple{prefix + "opml", webfw.MethodPost, "opml"},
	}
}

func (con Feed) Handler(c context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		action := webfw.GetMultiPatternIdentifier(c, r)
		user := readeef.GetUser(c, r)

		r.ParseForm()

		var resp responseError
		var feedId int64

		params := webfw.GetParams(c, r)

		if resp.err == nil {
			switch action {
			case "list":
				resp = listFeeds(user)
			case "discover":
				link := r.FormValue("url")
				resp = discoverFeeds(user, con.fm, link)
			case "opml":
				buf := util.BufferPool.GetBuffer()
				defer util.BufferPool.Put(buf)

				buf.ReadFrom(r.Body)

				resp = parseOpml(user, con.fm, buf.Bytes())
			case "add":
				links := r.Form["url"]
				resp = addFeeds(user, con.fm, links)
			case "remove":
				if feedId, resp.err = strconv.ParseInt(params["feed-id"], 10, 64); resp.err == nil {
					resp = removeFeed(user, con.fm, data.FeedId(feedId))
				}
			case "tags":
				if feedId, resp.err = strconv.ParseInt(params["feed-id"], 10, 64); resp.err == nil {
					if r.Method == "GET" {
						resp = getFeedTags(user, data.FeedId(feedId))
					} else if r.Method == "POST" {
						decoder := json.NewDecoder(r.Body)

						tags := []data.TagValue{}
						if resp.err = decoder.Decode(&tags); resp.err != nil && resp.err != io.EOF {
							break
						}

						resp.err = nil
						resp = setFeedTags(user, data.FeedId(feedId), tags)
					}
				}
			case "read":
				var timestamp int64

				if timestamp, resp.err = strconv.ParseInt(params["timestamp"], 10, 64); resp.err == nil {
					resp = markFeedAsRead(user, params["feed-id"], timestamp)
				}
			case "articles":
				var limit, offset int

				if limit, resp.err = strconv.Atoi(params["limit"]); resp.err == nil {
					if offset, resp.err = strconv.Atoi(params["offset"]); resp.err == nil {
						resp = getFeedArticles(user, con.sp, params["feed-id"], limit, offset,
							params["newer-first"] == "true", params["unread-only"] == "true")
					}
				}
			}
		}

		switch resp.err {
		case readeef.ErrNoAbsolute:
			resp.val["Error"] = true
			resp.val["ErrorType"] = errTypeNoAbsolute
			resp.err = nil
		case readeef.ErrNoFeed:
			resp.val["Error"] = true
			resp.val["ErrorType"] = errTypeNoFeed
			resp.err = nil
		}

		var b []byte
		if resp.err == nil {
			b, resp.err = json.Marshal(resp.val)
		}

		if resp.err == nil {
			w.Write(b)
		} else {
			webfw.GetLogger(c).Print(resp.err)

			w.WriteHeader(http.StatusInternalServerError)
		}

	})
}

func (con Feed) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}

func (p listFeedsProcessor) Process() responseError {
	return listFeeds(p.user)
}

func (p discoverFeedsProcessor) Process() responseError {
	return discoverFeeds(p.user, p.fm, p.Link)
}

func (p parseOpmlProcessor) Process() responseError {
	return parseOpml(p.user, p.fm, []byte(p.Opml))
}

func (p addFeedsProcessor) Process() responseError {
	return addFeeds(p.user, p.fm, p.Links)
}

func (p removeFeedProcessor) Process() responseError {
	return removeFeed(p.user, p.fm, p.Id)
}

func (p getFeedTagsProcessor) Process() responseError {
	return getFeedTags(p.user, p.Id)
}

func (p setFeedTagsProcessor) Process() responseError {
	return setFeedTags(p.user, p.Id, p.Tags)
}

func (p markFeedAsReadProcessor) Process() responseError {
	return markFeedAsRead(p.user, p.Id, p.Timestamp)
}

func (p getFeedArticlesProcessor) Process() responseError {
	return getFeedArticles(p.user, p.sp, p.Id, p.Limit, p.Offset, p.NewerFirst, p.UnreadOnly)
}

func listFeeds(user content.User) (resp responseError) {
	resp = newResponse()

	resp.val["Feeds"], resp.err = user.AllTaggedFeeds(), user.Err()
	return
}

func discoverFeeds(user content.User, fm *readeef.FeedManager, link string) (resp responseError) {
	resp = newResponse()

	if u, err := url.Parse(link); err != nil {
		resp.err = readeef.ErrNoAbsolute
		resp.errType = errTypeNoAbsolute
		return
	} else if !u.IsAbs() {
		u.Scheme = "http"
		if u.Host == "" {
			parts := strings.SplitN(u.Path, "/", 2)
			u.Host = parts[0]
			if len(parts) > 1 {
				u.Path = "/" + parts[1]
			} else {
				u.Path = ""
			}
		}
		link = u.String()
	}

	feeds, err := fm.DiscoverFeeds(link)
	if err != nil {
		resp.val["Feeds"] = []content.Feed{}
		return
	}

	uf := user.AllFeeds()
	if user.HasErr() {
		resp.err = user.Err()
		return
	}

	userFeedIdMap := make(map[data.FeedId]bool)
	userFeedLinkMap := make(map[string]bool)
	for i := range uf {
		in := uf[i].Data()
		userFeedIdMap[in.Id] = true
		userFeedLinkMap[in.Link] = true

		u, err := url.Parse(in.Link)
		if err == nil && strings.HasPrefix(u.Host, "www.") {
			u.Host = u.Host[4:]
			userFeedLinkMap[u.String()] = true
		}
	}

	respFeeds := []content.Feed{}
	for i := range feeds {
		in := feeds[i].Data()
		if !userFeedIdMap[in.Id] && !userFeedLinkMap[in.Link] {
			respFeeds = append(respFeeds, feeds[i])
		}
	}

	resp.val["Feeds"] = respFeeds
	return
}

func parseOpml(user content.User, fm *readeef.FeedManager, opmlData []byte) (resp responseError) {
	resp = newResponse()

	var opml parser.Opml
	if opml, resp.err = parser.ParseOpml(opmlData); resp.err != nil {
		return
	}

	uf := user.AllFeeds()
	if user.HasErr() {
		resp.err = user.Err()
		return
	}

	userFeedMap := make(map[data.FeedId]bool)
	for i := range uf {
		userFeedMap[uf[i].Data().Id] = true
	}

	var feeds []content.Feed
	for _, opmlFeed := range opml.Feeds {
		discovered, err := fm.DiscoverFeeds(opmlFeed.Url)
		if err != nil {
			continue
		}

		for _, f := range discovered {
			in := f.Data()
			if !userFeedMap[in.Id] {
				if len(opmlFeed.Tags) > 0 {
					in.Link += "#" + strings.Join(opmlFeed.Tags, ",")
				}

				f.Data(in)

				feeds = append(feeds, f)
			}
		}
	}

	resp.val["Feeds"] = feeds
	return
}

func addFeeds(user content.User, fm *readeef.FeedManager, links []string) (resp responseError) {
	resp = newResponse()

	success := false

	for _, link := range links {
		var u *url.URL
		if u, resp.err = url.Parse(link); resp.err != nil {
			/* TODO: non-fatal error */
			return
		} else if !u.IsAbs() {
			/* TODO: non-fatal error */
			resp.err = errors.New("Feed has no link")
			return
		} else {
			var f content.Feed
			if f, resp.err = fm.AddFeedByLink(link); resp.err != nil {
				return
			}

			uf := user.AddFeed(f)
			if uf.HasErr() {
				resp.err = f.Err()
				return
			}

			tags := strings.SplitN(u.Fragment, ",", -1)
			if u.Fragment != "" && len(tags) > 0 {
				repo := uf.Repo()
				tf := repo.TaggedFeed(user)
				tf.Data(uf.Data())

				t := make([]content.Tag, len(tags))
				for i := range tags {
					t[i] = repo.Tag(user)
					t[i].Value(data.TagValue(tags[i]))
				}

				tf.Tags(t)
				if tf.UpdateTags(); tf.HasErr() {
					resp.err = tf.Err()
					return
				}
			}

			success = true
		}
	}

	resp.val["Success"] = success
	return
}

func removeFeed(user content.User, fm *readeef.FeedManager, id data.FeedId) (resp responseError) {
	resp = newResponse()

	feed := user.FeedById(id)
	if feed.Delete(); feed.HasErr() {
		resp.err = feed.Err()
		return
	}

	fm.RemoveFeed(feed)

	resp.val["Success"] = true
	return
}

func getFeedTags(user content.User, id data.FeedId) (resp responseError) {
	resp = newResponse()

	repo := user.Repo()
	tf := repo.TaggedFeed(user)
	if uf := user.FeedById(id); uf.HasErr() {
		resp.err = uf.Err()
		return
	} else {
		tf.Data(uf.Data())
	}

	tags := tf.Tags()

	if tf.HasErr() {
		resp.err = tf.Err()
		return
	}

	t := make([]string, len(tags))
	for _, tag := range tags {
		t = append(t, tag.String())
	}

	resp.val["Tags"] = t
	return
}

func setFeedTags(user content.User, id data.FeedId, tagValues []data.TagValue) (resp responseError) {
	resp = newResponse()

	feed := user.FeedById(id)
	if feed.HasErr() {
		resp.err = feed.Err()
		return
	}

	repo := user.Repo()
	tf := repo.TaggedFeed(user)
	tf.Data(feed.Data())
	tags := make([]content.Tag, len(tagValues))
	for i := range tagValues {
		tags[i] = repo.Tag(user)
		tags[i].Value(tagValues[i])
	}
	tf.Tags(tags)
	if tf.UpdateTags(); tf.HasErr() {
		resp.err = tf.Err()
		return
	}

	resp.val["Success"] = true
	resp.val["Id"] = id
	return
}

func markFeedAsRead(user content.User, id string, timestamp int64) (resp responseError) {
	resp = newResponse()

	t := time.Unix(timestamp/1000, 0)

	switch {
	case id == "all":
		if user.ReadBefore(t, true); user.HasErr() {
			resp.err = user.Err()
			return
		}
	case id == "favorite" || strings.HasPrefix(id, "popular:"):
		// Favorites are assumbed to have been read already
	case strings.HasPrefix(id, "tag:"):
		tag := user.Repo().Tag(user)
		tag.Value(data.TagValue(id[4:]))
		if tag.ReadBefore(t, true); tag.HasErr() {
			resp.err = tag.Err()
			return
		}
	default:
		var feedId int64
		if feedId, resp.err = strconv.ParseInt(id, 10, 64); resp.err != nil {
			/* TODO: non-fatal error */
			return
		}

		feed := user.FeedById(data.FeedId(feedId))
		if feed.ReadBefore(t, true); feed.HasErr() {
			resp.err = feed.Err()
			return
		}
	}

	resp.val["Success"] = true
	return
}

func getFeedArticles(user content.User, sp content.SearchProvider, id string, limit int, offset int, newerFirst bool, unreadOnly bool) (resp responseError) {
	resp = newResponse()

	if limit > 200 {
		limit = 200
	}

	user.SortingByDate()
	if newerFirst {
		user.Order(data.DescendingOrder)
	} else {
		user.Order(data.AscendingOrder)
	}

	if id == "favorite" {
		resp.val["Articles"], resp.err = user.FavoriteArticles(limit, offset), user.Err()
	} else if id == "popular:all" {
		resp.val["Articles"], resp.err = user.ScoredArticles(time.Now().AddDate(0, 0, -5), time.Now(), limit, offset), user.Err()
	} else if id == "all" {
		if unreadOnly {
			resp.val["Articles"], resp.err = user.UnreadArticles(limit, offset), user.Err()
		} else {
			resp.val["Articles"], resp.err = user.Articles(limit, offset), user.Err()
		}
	} else if strings.HasPrefix(id, "popular:") {
		if strings.HasPrefix(id, "popular:tag:") {
			tag := user.Repo().Tag(user)
			tag.Value(data.TagValue(id[12:]))

			tag.SortingByDate()
			if newerFirst {
				tag.Order(data.DescendingOrder)
			} else {
				tag.Order(data.AscendingOrder)
			}
			resp.val["Articles"], resp.err = tag.ScoredArticles(time.Now().AddDate(0, 0, -5), time.Now(), limit, offset), tag.Err()
		} else {
			var f content.UserFeed

			var feedId int64
			feedId, resp.err = strconv.ParseInt(id[8:], 10, 64)

			if resp.err != nil {
				resp.err = errors.New("Unknown feed id " + id)
				return
			}

			if f = user.FeedById(data.FeedId(feedId)); f.HasErr() {
				/* TODO: non-fatal error */
				resp.err = f.Err()
				return
			}

			f.SortingByDate()
			if newerFirst {
				f.Order(data.DescendingOrder)
			} else {
				f.Order(data.AscendingOrder)
			}

			resp.val["Articles"], resp.err = f.ScoredArticles(time.Now().AddDate(0, 0, -5), time.Now(), limit, offset), f.Err()
		}
	} else if strings.HasPrefix(id, "search:") && sp != nil {
		var query string
		id = id[7:]
		parts := strings.Split(id, ":")

		if parts[0] == "tag" {
			id = strings.Join(parts[:2], ":")
			query = strings.Join(parts[2:], ":")
		} else {
			id = strings.Join(parts[:1], ":")
			query = strings.Join(parts[1:], ":")
		}

		sp.SortingByDate()
		if newerFirst {
			sp.Order(data.DescendingOrder)
		} else {
			sp.Order(data.AscendingOrder)
		}

		performSearch(&resp, user, sp, query, id, limit, offset)
	} else if strings.HasPrefix(id, "tag:") {
		tag := user.Repo().Tag(user)
		tag.Value(data.TagValue(id[4:]))

		tag.SortingByDate()
		if newerFirst {
			tag.Order(data.DescendingOrder)
		} else {
			tag.Order(data.AscendingOrder)
		}

		if unreadOnly {
			resp.val["Articles"], resp.err = tag.UnreadArticles(limit, offset), tag.Err()
		} else {
			resp.val["Articles"], resp.err = tag.Articles(limit, offset), tag.Err()
		}
	} else {
		var f content.UserFeed

		var feedId int64
		feedId, resp.err = strconv.ParseInt(id, 10, 64)

		if resp.err != nil {
			resp.err = errors.New("Unknown feed id " + id)
			return
		}

		if f = user.FeedById(data.FeedId(feedId)); f.HasErr() {
			/* TODO: non-fatal error */
			resp.err = f.Err()
			return
		}

		if newerFirst {
			f.Order(data.DescendingOrder)
		} else {
			f.Order(data.AscendingOrder)
		}

		f.SortingByDate()
		if unreadOnly {
			resp.val["Articles"], resp.err = f.UnreadArticles(limit, offset), f.Err()
		} else {
			resp.val["Articles"], resp.err = f.Articles(limit, offset), f.Err()
		}
	}

	switch articles := resp.val["Articles"].(type) {
	case []content.UserArticle:
		for _, a := range articles {
			insertThumbnail(a)
		}
	case []content.ScoredArticle:
		for _, a := range articles {
			insertThumbnail(a)
		}
	}

	resp.val["Limit"] = limit
	resp.val["Offset"] = offset

	return
}

func insertThumbnail(a content.Article) {
	d := a.Data()
	t := a.Thumbnail()

	d.Thumbnail = t.Base64DataUri()
	a.Data(d)
}

func performSearch(resp *responseError, user content.User, sp content.SearchProvider, query, feedId string, limit, offset int) {
	defer func() {
		if rec := recover(); rec != nil {
			resp.err = fmt.Errorf("Error during search: %s", rec)
		}
	}()

	if strings.HasPrefix(feedId, "tag:") {
		tag := user.Repo().Tag(user)
		tag.Value(data.TagValue(feedId[4:]))

		resp.val["Articles"], resp.err = tag.Query(query, sp, limit, offset), tag.Err()
	} else {
		if id, err := strconv.ParseInt(feedId, 10, 64); err == nil {
			f := user.FeedById(data.FeedId(id))
			resp.val["Articles"], resp.err = f.Query(query, sp, limit, offset), f.Err()
		} else {
			resp.val["Articles"], resp.err = user.Query(query, sp, limit, offset), user.Err()
		}
	}
}
