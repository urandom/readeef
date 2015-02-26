package postgres

import (
	"github.com/urandom/readeef/content/sql"
	"github.com/urandom/readeef/db"
	"github.com/urandom/webfw"
)

type Repo struct {
	*sql.Repo
}

func NewRepo(db *db.DB, logger webfw.Logger) *Repo {
	return &Repo{Repo: sql.NewRepo(db, logger)}
}

func init() {
	db.SetSQL("get_user_feeds", getUserFeeds)
	db.SetSQL("get_user_tag_feeds", getUserTagFeeds)
}

const (
	getUserFeeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds uf
WHERE f.id = uf.feed_id
	AND uf.user_login = $1
ORDER BY f.title COLLATE "default"
`
	getUserTagFeeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds_tags uft
WHERE f.id = uft.feed_id
	AND uft.user_login = $1 AND uft.tag = $2
ORDER BY f.title COLLATE "default"
`
)
