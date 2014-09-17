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
	id TEXT,
	feed_id INTEGER,
	title TEXT,
	description TEXT,
	link TEXT,
	date TIMESTAMP,

	PRIMARY KEY(id, feed_id),
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
	article_id TEXT,
	article_feed_id INTEGER,

	PRIMARY KEY(user_login, article_id, article_feed_id),
	FOREIGN KEY(user_login) REFERENCES users(login) ON DELETE CASCADE,
	FOREIGN KEY(article_id, article_feed_id) REFERENCES articles(id, feed_id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS users_articles_fav (
	user_login TEXT,
	article_id TEXT,
	article_feed_id INTEGER,

	PRIMARY KEY(user_login, article_id, article_feed_id),
	FOREIGN KEY(user_login) REFERENCES users(login) ON DELETE CASCADE,
	FOREIGN KEY(article_id, article_feed_id) REFERENCES articles(id, feed_id) ON DELETE CASCADE
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
}
