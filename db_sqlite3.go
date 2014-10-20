// +build sqlite3,cgo

package readeef

import _ "github.com/mattn/go-sqlite3"

const (
	sqlite3_get_user_feeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds uf
WHERE f.id = uf.feed_id
	AND uf.user_login = $1
ORDER BY f.title COLLATE NOCASE
`
	sqlite3_get_user_tag_feeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds_tags uft
WHERE f.id = uft.feed_id
	AND uft.user_login = $1 AND uft.tag = $2
ORDER BY f.title COLLATE NOCASE
`

	sqlite3_get_latest_feed_articles = `
SELECT a.feed_id, a.id, a.title, a.description, a.link, a.date, a.guid
FROM articles a
WHERE a.feed_id = $1 AND a.date > DATE('NOW', '-5 days')
`

	sqlite3_get_scored_user_articles = `
SELECT uf.feed_id, a.id, a.title, a.description, a.link, a.date, a.guid,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
INNER JOIN articles_scores asc
	ON a.id = asc.article_id
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND uf.user_login = ar.user_login
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND uf.user_login = af.user_login
WHERE a.date > DATE('NOW', '-5 days')
ORDER BY asc.score, a.date
LIMIT $2
OFFSET $3
`

	sqlite3_get_scored_user_articles_desc = `
SELECT uf.feed_id, a.id, a.title, a.description, a.link, a.date, a.guid,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
INNER JOIN articles_scores asc
	ON a.id = asc.article_id
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND uf.user_login = ar.user_login
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND uf.user_login = af.user_login
WHERE a.date > DATE('NOW', '-5 days')
ORDER BY asc.score DESC, a.date DESC
LIMIT $2
OFFSET $3
`
)

var (
	init_sql_sqlite3 = []string{`
PRAGMA foreign_keys = ON;`, `
PRAGMA journal_mode = WAL;`, `
CREATE TABLE IF NOT EXISTS readeef (
	db_version INTEGER
)`, `
CREATE TABLE IF NOT EXISTS users (
	login TEXT PRIMARY KEY,
	first_name TEXT,
	last_name TEXT,
	email TEXT,
	admin INTEGER DEFAULT 0,
	active INTEGER DEFAULT 1,
	profile_data TEXT,
	hash_type TEXT,
	salt TEXT,
	hash TEXT,
	md5_api TEXT
)`, `
CREATE TABLE IF NOT EXISTS feeds (
	id INTEGER PRIMARY KEY,
	link TEXT NOT NULL UNIQUE,
	title TEXT,
	description TEXT,
	hub_link TEXT,
	site_link TEXT,
	update_error TEXT,
	subscribe_error TEXT
)`, `
CREATE TABLE IF NOT EXISTS feed_images (
	id INTEGER PRIMARY KEY,
	feed_id INTEGER NOT NULL,
	title TEXT,
	url TEXT,
	width INTEGER,
	height INTEGER,

	FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS articles (
	id INTEGER PRIMARY KEY,
	feed_id INTEGER,
	link TEXT,
	guid TEXT,
	title TEXT,
	description TEXT,
	date TIMESTAMP,

	UNIQUE(feed_id, link),
	UNIQUE(feed_id, guid),
	FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS users_feeds (
	user_login TEXT,
	feed_id INTEGER,

	PRIMARY KEY(user_login, feed_id),
	FOREIGN KEY(user_login) REFERENCES users(login) ON DELETE CASCADE,
	FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS users_feeds_tags (
	user_login TEXT,
	feed_id INTEGER,
	tag TEXT,

	PRIMARY KEY(user_login, feed_id, tag),
	FOREIGN KEY(user_login, feed_id) REFERENCES users_feeds(user_login, feed_id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS users_articles_read (
	user_login TEXT,
	article_id INTEGER,

	PRIMARY KEY(user_login, article_id),
	FOREIGN KEY(user_login) REFERENCES users(login) ON DELETE CASCADE,
	FOREIGN KEY(article_id) REFERENCES articles(id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS users_articles_fav (
	user_login TEXT,
	article_id INTEGER,

	PRIMARY KEY(user_login, article_id),
	FOREIGN KEY(user_login) REFERENCES users(login) ON DELETE CASCADE,
	FOREIGN KEY(article_id) REFERENCES articles(id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS articles_scores (
	article_id BIGINT,
	score  INTEGER,
	score1 INTEGER,
	score2 INTEGER,
	score3 INTEGER,
	score4 INTEGER,
	score5 INTEGER,

	PRIMARY KEY(article_id),
	FOREIGN KEY(article_id) REFERENCES articles(id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS hubbub_subscriptions (
	feed_id INTEGER,
	link TEXT,
	lease_duration INTEGER,
	verification_time TIMESTAMP,
	subscription_failure INTEGER DEFAULT 0,

	PRIMARY KEY(feed_id),
	FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
)`,
	}
)

func init() {
	init_sql["sqlite3"] = init_sql_sqlite3
	sql_stmt["sqlite3:get_user_feeds"] = sqlite3_get_user_feeds
	sql_stmt["sqlite3:get_user_tag_feeds"] = sqlite3_get_user_tag_feeds
	sql_stmt["sqlite3:get_latest_feed_articles"] = sqlite3_get_latest_feed_articles
	sql_stmt["sqlite3:get_scored_user_articles"] = sqlite3_get_scored_user_articles
	sql_stmt["sqlite3:get_scored_user_articles_desc"] = sqlite3_get_scored_user_articles_desc
}
