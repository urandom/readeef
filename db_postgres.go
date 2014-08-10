// +build postgres

package readeef

import _ "github.com/lib/pq"

const (
	postgres_create_feed = `
INSERT INTO feeds(link, title, description, hub_link, update_error, subscribe_error)
	SELECT $1, $2, $3, $4, $5, $6 EXCEPT SELECT link, title, description, hub_link, update_error, subscribe_error FROM feeds WHERE link = $6 RETURNING id`
)

var (
	init_sql_postgres = []string{`
CREATE TABLE IF NOT EXISTS users (
	login TEXT PRIMARY KEY,
	first_name TEXT,
	last_name TEXT,
	email TEXT,
	hash_type TEXT,
	salt BYTEA,
	hash BYTEA,
	md5_api BYTEA
)`, `
CREATE TABLE IF NOT EXISTS feeds (
	id SERIAL PRIMARY KEY,
	link TEXT NOT NULL UNIQUE,
	title TEXT,
	description TEXT,
	hub_link TEXT,
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
	date TIMESTAMP WITH TIME ZONE,

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
	link TEXT,
	feed_id INTEGER NOT NULL UNIQUE,
	lease_duration INTEGER,
	verification_time TIMESTAMP WITH TIME ZONE,
	subscription_failure BOOLEAN,

	PRIMARY KEY(link),
	FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
)`,
	}
)

func init() {
	init_sql["postgres"] = init_sql_postgres
	sql_stmt["postgres:create_feed"] = postgres_create_feed
}
