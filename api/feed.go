package api

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
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

type exportOpmlProcessor struct {
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

type readStateProcessor struct {
	Id        string         `json:"id"`
	Timestamp int64          `json:"timestamp"`
	BeforeId  data.ArticleId `json:"beforeId"`

	user content.User
}

type getFeedArticlesProcessor struct {
	Id         string         `json:"id"`
	MinId      data.ArticleId `json:"minId"`
	MaxId      data.ArticleId `json:"maxId"`
	Limit      int            `json:"limit"`
	Offset     int            `json:"offset"`
	OlderFirst bool           `json:"olderFirst"`
	UnreadOnly bool           `json:"unreadOnly"`

	user content.User
	sp   content.SearchProvider
}

type addFeedError struct {
	Link  string
	Title string
	Error string
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
		webfw.MethodIdentifierTuple{prefix + ":feed-id/read/before/:before-id", webfw.MethodPost, "read"},
		webfw.MethodIdentifierTuple{prefix + ":feed-id/articles/:limit/:offset/:older-first/:unread-only", webfw.MethodGet, "articles"},
		webfw.MethodIdentifierTuple{prefix + ":feed-id/articles/min/:min-id/max/:max-id/:limit/:offset/:older-first/:unread-only", webfw.MethodGet, "articles"},

		webfw.MethodIdentifierTuple{prefix + "discover", webfw.MethodGet, "discover"},
		webfw.MethodIdentifierTuple{prefix + "opml", webfw.MethodGet, "opml-export"},
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

		switch action {
		case "list":
			resp = listFeeds(user)
		case "discover":
			link := r.FormValue("url")
			resp = discoverFeeds(user, con.fm, link)
		case "opml-export":
			resp = exportOpml(user)
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
					if b, err := ioutil.ReadAll(r.Body); err == nil {
						tags := []data.TagValue{}
						if err = json.Unmarshal(b, &tags); err != nil {
							resp.err = fmt.Errorf("Error decoding request body: %s", err)
							break
						}

						resp = setFeedTags(user, data.FeedId(feedId), tags)
					} else {
						resp.err = fmt.Errorf("Error reading request body: %s", err)
						break
					}
				}
			}
		case "read":
			var timestamp, beforeId int64

			if bid, ok := params["before-id"]; ok {
				beforeId, resp.err = strconv.ParseInt(bid, 10, 64)
			} else {
				timestamp, resp.err = strconv.ParseInt(params["timestamp"], 10, 64)
			}

			if resp.err == nil {
				resp = readState(user, params["feed-id"], data.ArticleId(beforeId), timestamp)
			}
		case "articles":
			var limit, offset int

			if limit, resp.err = strconv.Atoi(params["limit"]); resp.err == nil {
				if offset, resp.err = strconv.Atoi(params["offset"]); resp.err == nil {
					minId, _ := strconv.ParseInt(params["min-id"], 10, 64)
					maxId, _ := strconv.ParseInt(params["max-id"], 10, 64)

					resp = getFeedArticles(user, con.sp, params["feed-id"],
						data.ArticleId(minId), data.ArticleId(maxId),
						limit, offset, params["older-first"] == "true", params["unread-only"] == "true")
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

func (p exportOpmlProcessor) Process() responseError {
	return exportOpml(p.user)
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

func (p readStateProcessor) Process() responseError {
	return readState(p.user, p.Id, p.BeforeId, p.Timestamp)
}

func (p getFeedArticlesProcessor) Process() responseError {
	return getFeedArticles(p.user, p.sp, p.Id, p.MinId, p.MaxId,
		p.Limit, p.Offset, p.OlderFirst, p.UnreadOnly)
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

func exportOpml(user content.User) (resp responseError) {
	resp = newResponse()

	o := parser.OpmlXml{
		Version: "1.1",
		Head:    parser.OpmlHead{Title: "Feed subscriptions of " + user.String() + " from readeef"},
	}

	if feeds := user.AllTaggedFeeds(); user.HasErr() {
		resp.err = user.Err()
		return
	} else {
		body := parser.OpmlBody{}
		for _, f := range feeds {
			d := f.Data()

			tags := f.Tags()
			category := make([]string, len(tags))
			for i, t := range tags {
				category[i] = string(t.Data().Value)
			}
			body.Outline = append(body.Outline, parser.OpmlOutline{
				Text:     d.Title,
				Title:    d.Title,
				XmlUrl:   d.Link,
				HtmlUrl:  d.SiteLink,
				Category: strings.Join(category, ","),
				Type:     "rss",
			})
		}

		o.Body = body
	}

	var b []byte
	if b, resp.err = xml.MarshalIndent(o, "", "    "); resp.err != nil {
		return
	}

	resp.val["opml"] = xml.Header + string(b)

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

	var err error
	errs := make([]addFeedError, 0, len(links))

	for _, link := range links {
		var u *url.URL
		if u, err = url.Parse(link); err != nil {
			resp.err = err
			errs = append(errs, addFeedError{Link: link, Error: "Error parsing link"})
			continue
		} else if !u.IsAbs() {
			resp.err = errors.New("Feed has no link")
			errs = append(errs, addFeedError{Link: link, Error: resp.err.Error()})
			continue
		} else {
			var f content.Feed
			if f, err = fm.AddFeedByLink(link); err != nil {
				resp.err = err
				errs = append(errs, addFeedError{Link: link, Error: "Error adding feed to the database"})
				continue
			}

			uf := user.AddFeed(f)
			if uf.HasErr() {
				resp.err = f.Err()
				errs = append(errs, addFeedError{Link: link, Title: f.Data().Title, Error: "Error adding feed to the database"})
				continue
			}

			tags := strings.SplitN(u.Fragment, ",", -1)
			if u.Fragment != "" && len(tags) > 0 {
				repo := uf.Repo()
				tf := repo.TaggedFeed(user)
				tf.Data(uf.Data())

				t := make([]content.Tag, len(tags))
				for i := range tags {
					t[i] = repo.Tag(user)
					t[i].Data(data.Tag{Value: data.TagValue(tags[i])})
				}

				tf.Tags(t)
				if tf.UpdateTags(); tf.HasErr() {
					resp.err = tf.Err()
					errs = append(errs, addFeedError{Link: link, Title: f.Data().Title, Error: "Error adding feed to the database"})
					continue
				}
			}
		}
	}

	resp.val["Errors"] = errs
	resp.val["Success"] = len(errs) < len(links)
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

	filtered := make([]data.TagValue, 0, len(tagValues))
	for _, v := range tagValues {
		v = data.TagValue(strings.TrimSpace(string(v)))
		if v != "" {
			filtered = append(filtered, v)
		}
	}

	tags := make([]content.Tag, len(filtered))
	for i := range filtered {
		tags[i] = repo.Tag(user)
		tags[i].Data(data.Tag{Value: filtered[i]})
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

func readState(user content.User, id string, beforeId data.ArticleId, timestamp int64) (resp responseError) {
	resp = newResponse()

	var ar content.ArticleRepo

	o := data.ArticleUpdateStateOptions{}

	if timestamp > 0 {
		t := time.Unix(timestamp/1000, 0)
		o.BeforeDate = t
	}

	if beforeId > 0 {
		o.BeforeId = beforeId
	}

	switch {
	case id == "all":
		ar = user
	case id == "favorite":
		o.FavoriteOnly = true
		ar = user
	case strings.HasPrefix(id, "popular:"):
		// Can't bulk set state to popular articles
	case strings.HasPrefix(id, "tag:"):
		tag := user.Repo().Tag(user)
		tag.Data(data.Tag{Value: data.TagValue(id[4:])})

		ar = tag
	default:
		var feedId int64
		if feedId, resp.err = strconv.ParseInt(id, 10, 64); resp.err != nil {
			/* TODO: non-fatal error */
			return
		}

		ar = user.FeedById(data.FeedId(feedId))
	}

	if ar != nil {
		ar.ReadState(true, o)

		if e, ok := ar.(content.Error); ok && e.HasErr() {
			resp.err = e.Err()
			return
		}
	}

	resp.val["Success"] = true
	return
}

func getFeedArticles(user content.User, sp content.SearchProvider,
	id string, minId, maxId data.ArticleId, limit int, offset int, olderFirst bool,
	unreadOnly bool) (resp responseError) {

	resp = newResponse()

	if limit > 200 {
		limit = 200
	}

	var as content.ArticleSorting
	var ar content.ArticleRepo
	var ua []content.UserArticle

	o := data.ArticleQueryOptions{Limit: limit, Offset: offset, UnreadOnly: unreadOnly, UnreadFirst: true}

	if maxId > 0 {
		o.AfterId = maxId
		resp.val["MaxId"] = maxId
	}

	if id == "favorite" {
		o.FavoriteOnly = true
		ar = user
		as = user
	} else if id == "all" {
		ar = user
		as = user
	} else if strings.HasPrefix(id, "popular:") {
		o.IncludeScores = true
		o.HighScoredFirst = true
		o.BeforeDate = time.Now()
		o.AfterDate = time.Now().AddDate(0, 0, -5)

		if id == "popular:all" {
			ar = user
			as = user
		} else if strings.HasPrefix(id, "popular:tag:") {
			tag := user.Repo().Tag(user)
			tag.Data(data.Tag{Value: data.TagValue(id[12:])})

			ar = tag
			as = tag
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

			ar = f
			as = f
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
		if olderFirst {
			sp.Order(data.AscendingOrder)
		} else {
			sp.Order(data.DescendingOrder)
		}

		ua, resp.err = performSearch(user, sp, query, id, limit, offset)
	} else if strings.HasPrefix(id, "tag:") {
		tag := user.Repo().Tag(user)
		tag.Data(data.Tag{Value: data.TagValue(id[4:])})

		as = tag
		ar = tag
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

		as = f
		ar = f
	}

	if as != nil {
		as.SortingByDate()
		if olderFirst {
			as.Order(data.AscendingOrder)
		} else {
			as.Order(data.DescendingOrder)
		}
	}

	if ar != nil {
		ua = ar.Articles(o)

		if minId > 0 {
			qo := data.ArticleIdQueryOptions{BeforeId: maxId + 1, AfterId: minId - 1}

			qo.UnreadOnly = true
			resp.val["UnreadIds"] = ar.Ids(qo)

			qo.UnreadOnly = false
			qo.FavoriteOnly = true
			resp.val["FavoriteIds"] = ar.Ids(qo)

			resp.val["MinId"] = minId
		}

		if e, ok := ar.(content.Error); ok && e.HasErr() {
			resp.err = e.Err()
		}
	}

	resp.val["Articles"] = ua
	resp.val["Limit"] = limit
	resp.val["Offset"] = offset

	return
}

func performSearch(user content.User, sp content.SearchProvider, query, feedId string, limit, offset int) (ua []content.UserArticle, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("Error during search: %s", rec)
		}
	}()

	if strings.HasPrefix(feedId, "tag:") {
		tag := user.Repo().Tag(user)
		tag.Data(data.Tag{Value: data.TagValue(feedId[4:])})

		ua, err = tag.Query(query, sp, limit, offset), tag.Err()
	} else {
		if id, err := strconv.ParseInt(feedId, 10, 64); err == nil {
			f := user.FeedById(data.FeedId(id))
			ua, err = f.Query(query, sp, limit, offset), f.Err()
		} else {
			ua, err = user.Query(query, sp, limit, offset), user.Err()
		}
	}

	return
}
