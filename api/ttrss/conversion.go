package ttrss

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/urandom/readeef/content"
)

func convertRequest(in map[string]interface{}) request {
	req := request{}
	for key, v := range in {
		switch key {
		case "op":
			req.Op = parseString(v)
		case "sid":
			req.Sid = parseString(v)
		case "user":
			req.User = parseString(v)
		case "password":
			req.Password = parseString(v)
		case "output_mode":
			req.OutputMode = parseString(v)
		case "view_mode":
			req.ViewMode = parseString(v)
		case "order_by":
			req.OrderBy = parseString(v)
		case "search":
			req.Search = parseString(v)
		case "data":
			req.Data = parseString(v)
		case "pref_name":
			req.PrefName = parseString(v)
		case "feed_url":
			req.FeedUrl = parseString(v)
		case "unread_only":
			req.UnreadOnly = parseBool(v)
		case "include_empty":
			req.IncludeEmpty = parseBool(v)
		case "is_cat":
			req.IsCat = parseBool(v)
		case "show_content":
			req.ShowContent = parseBool(v)
		case "show_excerpt":
			req.ShowExcerpt = parseBool(v)
		case "sanitize":
			req.Sanitize = parseBool(v)
		case "has_sandbox":
			req.HasSandbox = parseBool(v)
		case "include_header":
			req.IncludeHeader = parseBool(v)
		case "seq":
			req.Seq = parseInt(v)
		case "limit":
			req.Limit = parseInt(v)
		case "offset":
			req.Offset = parseInt(v)
		case "skip":
			req.Skip = parseInt(v)
		case "mode":
			req.Mode = parseInt(v)
		case "field":
			req.Field = parseInt(v)
		case "cat_id":
			req.CatId = content.TagID(parseInt64(v))
		case "feed_id":
			req.FeedId = content.FeedID(parseInt64(v))
		case "since_id":
			req.SinceId = content.ArticleID(parseInt64(v))
		case "article_ids":
			req.ArticleIds = parseArticleIds(v)
		case "article_id":
			req.ArticleId = parseArticleIds(v)
		}
	}

	return req
}

func parseString(vv interface{}) string {
	if v, ok := vv.(string); ok {
		return v
	}
	return fmt.Sprintf("%v", vv)
}

func parseBool(vv interface{}) bool {
	switch v := vv.(type) {
	case string:
		return v == "t" || v == "true" || v == "1"
	case float64:
		return v == 1
	case bool:
		return v
	}
	return false
}

func parseInt(vv interface{}) int {
	switch v := vv.(type) {
	case string:
		i, _ := strconv.Atoi(v)
		return i
	case float64:
		return int(v)
	}
	return 0
}

func parseInt64(vv interface{}) int64 {
	switch v := vv.(type) {
	case string:
		i, _ := strconv.ParseInt(v, 10, 64)
		return i
	case float64:
		return int64(v)
	}
	return 0
}

func parseArticleIds(vv interface{}) (ids []content.ArticleID) {
	switch v := vv.(type) {
	case string:
		parts := strings.Split(v, ",")
		for _, p := range parts {
			if i, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64); err == nil {
				ids = append(ids, content.ArticleID(i))
			}
		}
	case []float64:
		for _, p := range v {
			ids = append(ids, content.ArticleID(int64(p)))
		}
	case float64:
		ids = append(ids, content.ArticleID(int64(v)))
	}
	return
}
