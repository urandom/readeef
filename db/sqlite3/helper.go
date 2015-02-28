package sqlite3

import (
	"github.com/urandom/readeef/db"
	"github.com/urandom/readeef/db/base"
)

type Helper struct {
	base.Helper
}

func (h Helper) InitSQL() []string {
	return initSQL
}

func init() {
	helper := &Helper{Helper: base.NewHelper()}

	helper.Set("get_user_feeds", getUserFeeds)
	helper.Set("get_user_tag_feeds", getUserTagFeeds)
	helper.Set("get_latest_feed_articles", getLatestFeedArticles)

	db.Register("sqlite3", helper)
}

const (
	getUserFeeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds uf
WHERE f.id = uf.feed_id
	AND uf.user_login = $1
ORDER BY f.title COLLATE NOCASE
`
	getUserTagFeeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds_tags uft
WHERE f.id = uft.feed_id
	AND uft.user_login = $1 AND uft.tag = $2
ORDER BY f.title COLLATE NOCASE
`

	getLatestFeedArticles = `
SELECT a.feed_id, a.id, a.title, a.description, a.link, a.date, a.guid
FROM articles a
WHERE a.feed_id = $1 AND a.date > DATE('NOW', '-5 days')
`
)
