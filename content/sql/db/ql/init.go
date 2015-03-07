package ql

var (
	initSQL = []string{`
CREATE TABLE IF NOT EXISTS readeef (
	db_version UINT32
)`, `
CREATE TABLE IF NOT EXISTS users (
	login STRING,
	first_name STRING,
	last_name STRING,
	email STRING,
	admin BOOL,
	active BOOL,
	profile_data BLOB,
	hash_type STRING,
	salt BLOB,
	hash BLOB,
	md5_api BLOB
)`, `
CREATE UNIQUE INDEX IF NOT EXISTS uniq_users_login
	ON users (login)
`, `
CREATE TABLE IF NOT EXISTS feeds (
	link STRING,
	title STRING,
	description STRING,
	hub_link STRING,
	site_link STRING,
	update_error STRING,
	subscribe_error STRING
)`, `
CREATE INDEX IF NOT EXISTS idx_feeds_id
	ON feeds (id())
)`, `
CREATE UNIQUE INDEX IF NOT EXISTS uniq_feeds_link
	ON feeds (link)
)`, `
CREATE TABLE IF NOT EXISTS feed_images (
	feed_id UINT64,
	title STRING,
	url STRING,
	width UINT32,
	height UINT32,
)`, `
CREATE TABLE IF NOT EXISTS articles (
	feed_id UINT64,
	link STRING,
	guid STRING,
	title STRING,
	description STRING,
	date time,

	_feed_id_link STRING,
	_feed_id_guid STRING
)`, `
CREATE INDEX IF NOT EXISTS idx_article_id
	ON articles (id())
)`, `
CREATE UNIQUE INDEX IF NOT EXISTS uniq_articles_feed_id_link
	ON articles (_feed_id_link)
)`, `
CREATE UNIQUE INDEX IF NOT EXISTS uniq_articles_feed_id_guid
	ON articles (_feed_id_guid)
)`, `
CREATE TABLE IF NOT EXISTS users_feeds (
	user_login STRING,
	feed_id UINT32,

	_login_id STRING
)`, `
CREATE UNIQUE INDEX IF NOT EXISTS uniq_users_feeds_login_id
	ON users_feeds (_login_id)
)`, `
CREATE TABLE IF NOT EXISTS users_feeds_tags (
	user_login STRING,
	feed_id UINT32,
	tag STRING,

	_login_id_tag STRING
)`, `
CREATE UNIQUE INDEX IF NOT EXISTS uniq_users_feeds_tags_login_id_tag
	ON users_feeds_tags (_login_id_tag)
)`, `
CREATE TABLE IF NOT EXISTS users_articles_read (
	user_login STRING,
	article_id UINT64,

	_login_id STRING
)`, `
CREATE UNIQUE INDEX IF NOT EXISTS uniq_users_articles_read_login_id
	ON users_articles_read (_login_id)
)`, `
CREATE TABLE IF NOT EXISTS users_articles_fav (
	user_login STRING,
	article_id UINT64,

	_login_id STRING
)`, `
CREATE UNIQUE INDEX IF NOT EXISTS uniq_users_articles_fav_login_id
	ON users_articles_fav (_login_id)
)`, `
CREATE TABLE IF NOT EXISTS articles_scores (
	article_id UINT64,
	score  UINT64,
	score1 UINT64,
	score2 UINT64,
	score3 UINT64,
	score4 UINT64,
	score5 UINT64
)`, `
CREATE UNIQUE INDEX IF NOT EXISTS uniq_articles_scores_article_id
	ON articles_scores (article_id)
)`, `
CREATE TABLE IF NOT EXISTS hubbub_subscriptions (
	feed_id UINT32,
	link STRING,
	lease_duration UINT32,
	verification_time TIME,
	subscription_failure BOOL
)`, `
CREATE UNIQUE INDEX IF NOT EXISTS uniq_hubbub_subscriptions_feed_id
	ON hubbub_subscriptions (feed_id)
)`,
	}
)
