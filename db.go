package readeef

import (
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var (
	init_sql = []string{`
PRAGMA foreign_keys = ON;`, `
CREATE TABLE IF NOT EXISTS users (
	login TEXT PRIMARY KEY,
	first_name TEXT,
	last_name TEXT,
	email TEXT,
	salt TEXT,
	hash TEXT,
	md5_api TEXT
);`, `
CREATE TABLE IF NOT EXISTS feeds (
	id INTEGER PRIMARY KEY,	
	title TEXT,
	description TEXT,
	link TEXT,
	hub_link TEXT
)`, `
CREATE TABLE IF NOT EXISTS feed_images (
	id INTEGER PRIMARY KEY,	
	feed_id INTEGER,
	title TEXT,
	url TEXT,
	width INTEGER,
	height INTEGER,

	FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS articles (
	id INTEGER PRIMARY KEY,	
	feed_id INTEGER,
	title TEXT,
	description TEXT,
	link TEXT,
	date DATETIME,

	FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS users_feeds (
	user_login TEXT,
	feed_id INTEGER,

	FOREIGN KEY(user_login) REFERENCES user(login) ON DELETE CASCADE,
	FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS users_articles_read (
	user_login TEXT,
	article_id INTEGER,

	FOREIGN KEY(user_login) REFERENCES user(login) ON DELETE CASCADE,
	FOREIGN KEY(article_id) REFERENCES articles(id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS users_articles_fav (
	user_login TEXT,
	article_id INTEGER,

	FOREIGN KEY(user_login) REFERENCES user(login) ON DELETE CASCADE,
	FOREIGN KEY(article_id) REFERENCES articles(id) ON DELETE CASCADE
)`,
	}
)

type Validator interface {
	Validate() error
}

type DB struct {
	*sqlx.DB
	driver        string
	connectString string
}

type ValidationError error

func NewDB(driver, conn string) DB {
	return DB{driver: driver, connectString: conn}
}

func (db *DB) Connect() error {
	dbx, err := sqlx.Connect(db.driver, db.connectString)
	if err != nil {
		return err
	}

	db.DB = dbx

	return db.init()
}

func (db DB) init() error {
	for _, sql := range init_sql {
		_, err := db.Exec(sql)
		if err != nil {
			return errors.New(fmt.Sprintf("Error executing '%s': %v", sql, err))
		}
	}

	return nil
}
