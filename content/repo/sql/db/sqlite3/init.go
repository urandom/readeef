package sqlite3

var (
	initSQL = []string{`
PRAGMA foreign_keys = ON`, `
PRAGMA journal_mode = WAL`, `
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
	salt BLOB,
	hash BLOB,
	md5_api BLOB
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
CREATE TABLE IF NOT EXISTS tags (
	id INTEGER PRIMARY KEY,
	value TEXT NOT NULL UNIQUE
)`, `
CREATE TABLE IF NOT EXISTS users_feeds_tags (
	user_login TEXT,
	feed_id INTEGER,
	tag_id INTEGER,

	PRIMARY KEY(user_login, feed_id, tag_id),
	FOREIGN KEY(user_login) REFERENCES users(login) ON DELETE CASCADE,
	FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
	FOREIGN KEY(tag_id) REFERENCES tags(id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS users_articles_unread (
	user_login TEXT,
	article_id BIGINT,
	insert_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

	PRIMARY KEY(user_login, article_id),
	FOREIGN KEY(user_login) REFERENCES users(login) ON DELETE CASCADE,
	FOREIGN KEY(article_id) REFERENCES articles(id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS users_articles_favorite (
	user_login TEXT,
	article_id BIGINT,

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
CREATE TABLE IF NOT EXISTS articles_thumbnails (
	article_id BIGINT,
	thumbnail TEXT NOT NULL DEFAULT '',
	link TEXT,
	processed INTEGER DEFAULT 0,

	PRIMARY KEY(article_id),
	FOREIGN KEY(article_id) REFERENCES articles(id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS articles_extracts (
	article_id BIGINT,
	title TEXT,
	content TEXT,
	top_image TEXT,
	language TEXT,

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
)`, `
CREATE INDEX IF NOT EXISTS articles_feed_id_idx ON articles (feed_id);
`, `
CREATE INDEX IF NOT EXISTS articles_title_idx ON articles (LOWER(title));
`, `
CREATE INDEX IF NOT EXISTS articles_link_idx ON articles (LOWER(link));
`, `
CREATE INDEX IF NOT EXISTS articles_date_idx ON articles (date);
`,
	}
)
