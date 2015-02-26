package ql

import "github.com/cznic/ql"

var (
	initSQL = []string{`
CREATE TABLE IF NOT EXISTS readeef (
	db_version INT
)`, `
CREATE TABLE IF NOT EXISTS users (
	login STRING PRIMARY KEY,
	first_name STRING,
	last_name STRING,
	email STRING,
	admin BOOL,
	active BOOL,
	profile_data BYTEA,
	hash_type STRING,
	salt BYTEA,
	hash BYTEA,
	md5_api BYTEA
)`, `
CREATE TABLE IF NOT EXISTS feeds (
	link STRING NOT NULL UNIQUE,
	title STRING,
	description STRING,
	hub_link STRING,
	site_link STRING,
	update_error STRING,
	subscribe_error STRING
)`, `
CREATE TABLE IF NOT EXISTS feed_images (
	feed_id INT NOT NULL,
	title STRING,
	url STRING,
	width INT,
	height INT,
)`, `
CREATE TABLE IF NOT EXISTS articles (
	feed_id INT,
	link STRING,
	guid STRING,
	title STRING,
	description STRING,
	date TIME,

	UNIQUE(feed_id, link),
	UNIQUE(feed_id, guid),
)`, `
CREATE TABLE IF NOT EXISTS users_feeds (
	user_login STRING,
	feed_id INT,

	PRIMARY KEY(user_login, feed_id),
)`, `
CREATE TABLE IF NOT EXISTS users_feeds_tags (
	user_login STRING,
	feed_id INT,
	tag STRING,

	PRIMARY KEY(user_login, feed_id, tag),
)`, `
CREATE TABLE IF NOT EXISTS users_articles_read (
	user_login STRING,
	article_id INT64,

	PRIMARY KEY(user_login, article_id),
)`, `
CREATE TABLE IF NOT EXISTS users_articles_fav (
	user_login STRING,
	article_id INT64,

	PRIMARY KEY(user_login, article_id),
)`, `
CREATE TABLE IF NOT EXISTS articles_scores (
	article_id INT64,
	score  INT64,
	score1 INT64,
	score2 INT64,
	score3 INT64,
	score4 INT64,
	score5 INT64,

	PRIMARY KEY(article_id),
)`, `
CREATE TABLE IF NOT EXISTS hubbub_subscriptions (
	feed_id INT,
	link STRING,
	lease_duration INT,
	verification_time TIME,
	subscription_failure BOOL,

	PRIMARY KEY(feed_id),
)`,
	}
)

func init() {
	ql.RegisterDriver()
}
