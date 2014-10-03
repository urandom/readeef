package readeef

import (
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	get_feed    = `SELECT link, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds WHERE id = $1`
	create_feed = `
INSERT INTO feeds(link, title, description, hub_link, site_link, update_error, subscribe_error)
	SELECT $1, $2, $3, $4, $5, $6, $7 EXCEPT SELECT link, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds WHERE link = $1`
	update_feed = `UPDATE feeds SET link = $1, title = $2, description = $3, hub_link = $4, site_link = $5, update_error = $6, subscribe_error = $7 WHERE id = $8`
	delete_feed = `DELETE FROM feeds WHERE id = $1`

	get_feed_by_link = `SELECT id, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds WHERE link = $1`

	get_feeds = `SELECT id, link, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds`

	get_unsubscribed_feeds = `
SELECT f.id, f.link, f.title, f.description, f.hub_link, f.site_link, f.update_error, f.subscribe_error
	FROM feeds f LEFT OUTER JOIN hubbub_subscriptions hs
	ON f.id = hs.feed_id AND hs.subscription_failure = '1'
	ORDER BY f.title
`

	get_user_feed = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds uf
WHERE f.id = uf.feed_id
	AND f.id = $1 AND uf.user_login = $2
`

	create_user_feed = `
INSERT INTO users_feeds(user_login, feed_id)
	SELECT $1, $2 EXCEPT SELECT user_login, feed_id FROM users_feeds WHERE user_login = $1 AND feed_id = $2`
	delete_user_feed = `DELETE FROM users_feeds WHERE user_login = $1 AND feed_id = $2`

	get_user_feeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds uf
WHERE f.id = uf.feed_id
	AND uf.user_login = $1
ORDER BY LOWER(f.title)
`

	get_feed_article = `
SELECT uf.feed_id, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.feed_id = $1 AND a.Id = $2
	AND uf.user_login = $3
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_id = ar.article_feed_id
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_id = af.article_feed_id
`

	update_feed_article = `
UPDATE articles SET title = $1, description = $2, link = $3, date = $4 WHERE id = $5 AND feed_id = $6
`

	get_feed_articles = `
SELECT uf.feed_id, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.feed_id = $1 AND uf.user_login = $2
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_id = ar.article_feed_id
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_id = af.article_feed_id
ORDER BY read, a.date
LIMIT $3
OFFSET $4
`

	get_feed_articles_desc = `
SELECT uf.feed_id, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.feed_id = $1 AND uf.user_login = $2
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_id = ar.article_feed_id
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_id = af.article_feed_id
ORDER BY read ASC, a.date DESC
LIMIT $3
OFFSET $4
`

	get_unread_feed_articles = `
SELECT uf.feed_id, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.feed_id = $1 AND uf.user_login = $2
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_id = ar.article_feed_id
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_id = af.article_feed_id
WHERE ar.article_id IS NULL
ORDER BY a.date
LIMIT $3
OFFSET $4
`

	get_unread_feed_articles_desc = `
SELECT uf.feed_id, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.feed_id = $1 AND uf.user_login = $2
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_id = ar.article_feed_id
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_id = af.article_feed_id
WHERE ar.article_id IS NULL
ORDER BY a.date DESC
LIMIT $3
OFFSET $4
`

	get_read_feed_articles = `
SELECT uf.feed_id, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.feed_id = $1 AND uf.user_login = $2
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_id = ar.article_feed_id
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_id = af.article_feed_id
WHERE ar.article_id IS NOT NULL
ORDER BY a.date
LIMIT $3
OFFSET $4
`

	get_user_articles = `
SELECT uf.feed_id, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_id = ar.article_feed_id
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_id = af.article_feed_id
ORDER BY read, a.date
LIMIT $2
OFFSET $3
`

	get_user_articles_desc = `
SELECT uf.feed_id, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_id = ar.article_feed_id
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_id = af.article_feed_id
ORDER BY read ASC, a.date DESC
LIMIT $2
OFFSET $3
`

	get_unread_user_articles = `
SELECT uf.feed_id, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_id = ar.article_feed_id
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_id = af.article_feed_id
WHERE ar.article_id IS NULL
ORDER BY a.date
LIMIT $2
OFFSET $3
`

	get_unread_user_articles_desc = `
SELECT uf.feed_id, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_id = ar.article_feed_id
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_id = af.article_feed_id
WHERE ar.article_id IS NULL
ORDER BY a.date DESC
LIMIT $2
OFFSET $3
`

	get_read_user_articles = `
SELECT uf.feed_id, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_id = ar.article_feed_id
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_id = af.article_feed_id
WHERE ar.article_id IS NOT NULL
ORDER BY a.date
LIMIT $2
OFFSET $3
`

	get_all_articles = `
SELECT a.feed_id, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM articles a
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_id = ar.article_feed_id
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_id = af.article_feed_id
`

	get_all_feed_articles = `
SELECT a.feed_id, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM articles a
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_id = ar.article_feed_id
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_id = af.article_feed_id
WHERE a.feed_id = $1
`

	create_all_user_articles_read_by_date = `
INSERT INTO users_articles_read
	SELECT uf.user_login, a.id, uf.feed_id
	FROM users_feeds uf INNER JOIN articles a
		ON uf.feed_id = a.feed_id AND uf.user_login = $1
		AND a.id IN (SELECT id FROM articles WHERE date IS NULL OR date < $2)
`

	delete_all_user_articles_read_by_date = `
DELETE FROM users_articles_read WHERE user_login = $1 AND article_id IN (
	SELECT id FROM articles WHERE date IS NULL OR date < $2
)
`

	create_all_users_articles_read_by_feed_date = `
INSERT INTO users_articles_read
	SELECT uf.user_login, a.id, uf.feed_id
	FROM users_feeds uf INNER JOIN articles a
		ON uf.feed_id = a.feed_id AND uf.user_login = $1 AND uf.feed_id = $2
		AND a.id IN (SELECT id FROM articles WHERE date IS NULL OR date < $3)
`

	delete_all_users_articles_read_by_feed_date = `
DELETE FROM users_articles_read WHERE user_login = $1 AND article_feed_id = $2 AND article_id IN (
	SELECT id FROM articles WHERE date IS NULL OR date < $3
)
`

	get_user_favorite_articles = `
SELECT uf.feed_id, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
1 AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
INNER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_id = af.article_feed_id
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_id = ar.article_feed_id
ORDER BY a.date
LIMIT $2
OFFSET $3
`

	get_user_favorite_articles_desc = `
SELECT uf.feed_id, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
1 AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
INNER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_id = af.article_feed_id
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_id = ar.article_feed_id
ORDER BY a.date DESC
LIMIT $2
OFFSET $3
`
)

var (
	ErrNoFeedUser = errors.New("Feed does not have an associated user.")
)

type FeedArticleIds struct {
	FeedId    int64
	ArticleId string
}

func (db DB) GetFeed(id int64) (Feed, error) {
	Debug.Printf("Getting feed for %d\n", id)

	var f Feed
	if err := db.Get(&f, db.NamedSQL("get_feed"), id); err != nil {
		return f, err
	}

	f.Id = id

	return f, nil
}

func (db DB) UpdateFeed(f Feed) (Feed, bool, error) {
	if err := f.Validate(); err != nil {
		return f, false, err
	}

	// FIXME: Remove when the 'FOREIGN KEY constraing failed' bug is removed
	if db.driver == "sqlite3" {
		db.Query("SELECT 1")
	}

	tx, err := db.Beginx()
	if err != nil {
		return f, false, err
	}
	defer tx.Rollback()

	Debug.Println("Updading feed " + f.Link)

	ustmt, err := tx.Preparex(db.NamedSQL("update_feed"))
	if err != nil {
		return f, false, err
	}
	defer ustmt.Close()

	res, err := ustmt.Exec(f.Link, f.Title, f.Description, f.HubLink, f.SiteLink, f.UpdateError, f.SubscribeError, f.Id)
	if err != nil {
		return f, false, err
	}

	if num, err := res.RowsAffected(); err != nil || num == 0 {
		cstmt, err := tx.Preparex(db.NamedSQL("create_feed"))
		if err != nil {
			return f, false, err
		}
		defer cstmt.Close()

		id, err := db.CreateWithId(cstmt, f.Link, f.Title, f.Description, f.HubLink, f.SiteLink, f.UpdateError, f.SubscribeError)
		if err != nil {
			return f, false, err
		}

		f.Id = id

		for i := 0; i < len(f.Articles); i++ {
			f.Articles[i].FeedId = id
		}
	}

	f.lastUpdatedArticleIds = map[string]bool{}
	newArticles, err := db.updateFeedArticles(tx, f, f.Articles)
	if err != nil {
		return f, false, err
	}

	tx.Commit()

	for _, a := range newArticles {
		f.lastUpdatedArticleIds[a.Id] = true
	}

	return f, len(newArticles) > 0, nil
}

func (db DB) DeleteFeed(f Feed) error {
	if err := f.Validate(); err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.NamedSQL("delete_feed"))

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(f.Id)
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func (db DB) GetFeedByLink(link string) (Feed, error) {
	Debug.Println("Getting feed " + link)

	var f Feed
	if err := db.Get(&f, db.NamedSQL("get_feed_by_link"), link); err != nil {
		return f, err
	}

	f.Link = link

	return f, nil
}

func (db DB) GetFeeds() ([]Feed, error) {
	var feeds []Feed

	if err := db.Select(&feeds, db.NamedSQL("get_feeds")); err != nil {
		return feeds, err
	}

	return feeds, nil
}

func (db DB) GetUnsubscribedFeed() ([]Feed, error) {
	var feeds []Feed

	if err := db.Select(&feeds, db.NamedSQL("get_unsubscribed_feeds")); err != nil {
		return feeds, err
	}

	return feeds, nil
}

func (db DB) GetUserFeed(id int64, u User) (Feed, error) {
	var f Feed

	if err := db.Get(&f, db.NamedSQL("get_user_feed"), id, u.Login); err != nil {
		return f, err
	}

	f.User = u

	return f, nil
}

func (db DB) CreateUserFeed(u User, f Feed) (Feed, error) {
	if err := u.Validate(); err != nil {
		return f, err
	}
	if err := f.Validate(); err != nil {
		return f, err
	}

	// FIXME: Remove when the 'FOREIGN KEY constraing failed' bug is removed
	if db.driver == "sqlite3" {
		db.Query("SELECT 1")
	}

	tx, err := db.Beginx()
	if err != nil {
		return f, err
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.NamedSQL("create_user_feed"))

	if err != nil {
		return f, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Login, f.Id)
	if err != nil {
		return f, err
	}

	tx.Commit()

	f.User = u

	return f, nil
}

func (db DB) DeleteUserFeed(f Feed) error {
	if err := f.Validate(); err != nil {
		return err
	}

	if err := f.User.Validate(); err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.NamedSQL("delete_user_feed"))

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(f.User.Login, f.Id)
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func (db DB) GetUserFeeds(u User) ([]Feed, error) {
	var feeds []Feed

	if err := db.Select(&feeds, db.NamedSQL("get_user_feeds"), u.Login); err != nil {
		return feeds, err
	}

	for _, f := range feeds {
		f.User = u
	}

	return feeds, nil
}

func (db DB) GetFeedArticle(feedId int64, articleId string, user User) (Article, error) {
	Debug.Printf("Getting feed article for %d - %s\n", feedId, articleId)

	var a Article
	if err := db.Get(&a, db.NamedSQL("get_feed_article"), feedId, articleId, user.Login); err != nil {
		return a, err
	}

	return a, nil
}

func (db DB) CreateFeedArticles(f Feed, articles []Article) (Feed, error) {
	if len(articles) == 0 {
		return f, nil
	}

	tx, err := db.Beginx()
	if err != nil {
		return f, err
	}
	defer tx.Rollback()

	if _, err := db.updateFeedArticles(tx, f, articles); err != nil {
		return f, err
	}

	for _, a := range articles {
		a.FeedId = f.Id
	}

	f.Articles = append(f.Articles, articles...)

	tx.Commit()

	return f, nil
}

func (db DB) GetFeedArticles(f Feed, paging ...int) (Feed, error) {
	return db.getFeedArticles(f, "get_feed_articles", paging...)
}

func (db DB) GetFeedArticlesDesc(f Feed, paging ...int) (Feed, error) {
	return db.getFeedArticles(f, "get_feed_articles_desc", paging...)
}

func (db DB) GetUnreadFeedArticles(f Feed, paging ...int) (Feed, error) {
	return db.getFeedArticles(f, "get_unread_feed_articles", paging...)
}

func (db DB) GetUnreadFeedArticlesDesc(f Feed, paging ...int) (Feed, error) {
	return db.getFeedArticles(f, "get_unread_feed_articles_desc", paging...)
}

func (db DB) GetReadFeedArticles(f Feed, paging ...int) (Feed, error) {
	if f.User.Login == "" {
		return f, ErrNoFeedUser
	}

	var articles []Article

	limit, offset := pagingLimit(paging)

	if err := db.Select(&articles, db.NamedSQL("get_read_feed_articles"), f.Id, f.User.Login, limit, offset); err != nil {
		return f, err
	}

	f.Articles = articles

	return f, nil
}

func (db DB) GetUserArticles(u User, paging ...int) ([]Article, error) {
	return db.getUserArticles(u, "get_user_articles", paging...)
}

func (db DB) GetUserArticlesDesc(u User, paging ...int) ([]Article, error) {
	return db.getUserArticles(u, "get_user_articles_desc", paging...)
}

func (db DB) GetUnreadUserArticles(u User, paging ...int) ([]Article, error) {
	return db.getUserArticles(u, "get_unread_user_articles", paging...)
}

func (db DB) GetUnreadUserArticlesDesc(u User, paging ...int) ([]Article, error) {
	return db.getUserArticles(u, "get_unread_user_articles_desc", paging...)
}

func (db DB) GetReadUserArticles(u User, paging ...int) ([]Article, error) {
	var articles []Article

	limit, offset := pagingLimit(paging)

	if err := db.Select(&articles, db.NamedSQL("get_read_user_articles"), u.Login, limit, offset); err != nil {
		return articles, err
	}

	return articles, nil
}

func (db DB) GetAllArticles() ([]Article, error) {
	var articles []Article

	if err := db.Select(&articles, db.NamedSQL("get_all_articles")); err != nil {
		return articles, err
	}

	return articles, nil
}

func (db DB) GetAllFeedArticles(f Feed) ([]Article, error) {
	var articles []Article

	if err := db.Select(&articles, db.NamedSQL("get_all_feed_articles"), f.Id); err != nil {
		return articles, err
	}

	return articles, nil
}

func (db DB) GetSpecificArticles(idTuples ...FeedArticleIds) ([]Article, error) {
	if len(idTuples) == 0 {
		return []Article{}, nil
	}

	sql := db.NamedSQL("get_all_articles")
	sql += ` WHERE `

	var args []interface{}
	index := 1
	for i, ids := range idTuples {
		if i != 0 {
			sql += ` OR `
		}

		sql += fmt.Sprintf(`(a.id = $%d AND a.feed_id = $%d)`, index, index+1)
		args = append(args, ids.ArticleId, ids.FeedId)
		index = len(args) + 1
	}

	var articles []Article

	if err := db.Select(&articles, sql, args...); err != nil {
		return articles, err
	}

	return articles, nil
}

func (db DB) MarkUserArticlesAsRead(u User, articles []Article, read bool) error {
	if len(articles) == 0 {
		return nil
	}

	var sql string
	var args []interface{}

	if read {
		sql = `INSERT INTO users_articles_read(user_login, article_id, article_feed_id) `
	} else {
		sql = `DELETE FROM users_articles_read WHERE `
	}

	index := 1
	for i, a := range articles {
		if err := a.Validate(); err != nil {
			return err
		}

		if read {
			if i != 0 {
				sql += ` UNION `
			}

			sql += fmt.Sprintf(`SELECT $%d, $%d, $%d EXCEPT SELECT user_login, article_id, article_feed_id FROM users_articles_read WHERE user_login = $%d AND article_id = $%d AND article_feed_id = $%d`, index, index+1, index+2, index, index+1, index+2)
			args = append(args, u.Login, a.Id, a.FeedId)
		} else {
			if i != 0 {
				sql += `OR `
			}

			sql += fmt.Sprintf(`(user_login = $%d AND article_id = $%d AND article_feed_id = $%d)`, index, index+1, index+2)
			args = append(args, u.Login, a.Id, a.FeedId)
		}
		index = len(args) + 1
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(sql)

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func (db DB) MarkUserArticlesByDateAsRead(u User, d time.Time, read bool) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.NamedSQL("delete_all_user_articles_read_by_date"))

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Login, d)
	if err != nil {
		return err
	}

	stmt, err = tx.Preparex(db.NamedSQL("create_all_user_articles_read_by_date"))

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Login, d)
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func (db DB) MarkFeedArticlesByDateAsRead(f Feed, d time.Time, read bool) error {
	if f.User.Login == "" {
		return ErrNoFeedUser
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.NamedSQL("delete_all_users_articles_read_by_feed_date"))

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(f.User.Login, f.Id, d)
	if err != nil {
		return err
	}

	stmt, err = tx.Preparex(db.NamedSQL("create_all_users_articles_read_by_feed_date"))

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(f.User.Login, f.Id, d)
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}
func (db DB) GetUserFavoriteArticles(u User, paging ...int) ([]Article, error) {
	return db.getUserFavoriteArticles(u, "get_user_favorite_articles", paging...)
}

func (db DB) GetUserFavoriteArticlesDesc(u User, paging ...int) ([]Article, error) {
	return db.getUserFavoriteArticles(u, "get_user_favorite_articles_desc", paging...)
}

func (db DB) MarkUserArticlesAsFavorite(u User, articles []Article, read bool) error {
	if len(articles) == 0 {
		return nil
	}

	var sql string
	var args []interface{}

	if read {
		sql = `INSERT INTO users_articles_fav(user_login, article_id, article_feed_id) `
	} else {
		sql = `DELETE FROM users_articles_fav WHERE `
	}

	index := 1
	for i, a := range articles {
		if err := a.Validate(); err != nil {
			return err
		}

		if read {
			if i != 0 {
				sql += ` UNION `
			}

			sql += fmt.Sprintf(`SELECT $%d, $%d, $%d EXCEPT SELECT user_login, article_id, article_feed_id FROM users_articles_fav WHERE user_login = $%d AND article_id = $%d AND article_feed_id = $%d`, index, index+1, index+2, index, index+1, index+2)
			args = append(args, u.Login, a.Id, a.FeedId)
		} else {
			if i != 0 {
				sql += `OR `
			}

			sql += fmt.Sprintf(`(user_login = $%d AND article_id = $%d AND article_feed_id = $%d)`, index, index+1, index+2)
			args = append(args, u.Login, a.Id, a.FeedId)
		}
		index = len(args) + 1
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(sql)

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func (db DB) updateFeedArticles(tx *sqlx.Tx, f Feed, articles []Article) ([]Article, error) {
	if len(articles) == 0 {
		return []Article{}, nil
	}

	Debug.Println("Updating feed articles for " + f.Link)

	newArticles := []Article{}

	for _, a := range articles {
		if err := a.Validate(); err != nil {
			return newArticles, err
		}

		stmt, err := tx.Preparex(db.NamedSQL("update_feed_article"))
		if err != nil {
			return newArticles, err
		}
		defer stmt.Close()

		res, err := stmt.Exec(a.Title, a.Description, a.Link, a.Date, a.Id, f.Id)
		if err != nil {
			return newArticles, err
		}

		if num, err := res.RowsAffected(); err != nil || num == 0 {
			newArticles = append(newArticles, a)
		}
	}

	if len(newArticles) == 0 {
		return newArticles, nil
	}

	sql := `INSERT INTO articles(id, feed_id, title, description, link, date) `
	args := []interface{}{}
	index := 1

	for i, a := range newArticles {
		if err := a.Validate(); err != nil {
			return newArticles, err
		}

		if i != 0 {
			sql += ` UNION `
		}

		sql += fmt.Sprintf(`SELECT $%d, $%d, $%d, $%d ,$%d, $%d EXCEPT SELECT id, feed_id, title, description, link, date FROM articles WHERE id = $%d AND feed_id = $%d`, index, index+1, index+2, index+3, index+4, index+5, index, index+1)
		args = append(args, a.Id, f.Id, a.Title, a.Description, a.Link, a.Date)
		index = len(args) + 1
	}

	stmt, err := tx.Preparex(sql)

	if err != nil {
		return newArticles, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return newArticles, err
	}

	return newArticles, nil
}

func (db DB) getFeedArticles(f Feed, namedSQL string, paging ...int) (Feed, error) {
	if f.User.Login == "" {
		return f, ErrNoFeedUser
	}

	var articles []Article

	limit, offset := pagingLimit(paging)

	if err := db.Select(&articles, db.NamedSQL(namedSQL), f.Id, f.User.Login, limit, offset); err != nil {
		return f, err
	}

	f.Articles = articles

	return f, nil
}

func (db DB) getUserArticles(u User, namedSQL string, paging ...int) ([]Article, error) {
	var articles []Article

	limit, offset := pagingLimit(paging)

	if err := db.Select(&articles, db.NamedSQL(namedSQL), u.Login, limit, offset); err != nil {
		return articles, err
	}

	return articles, nil
}

func (db DB) getUserFavoriteArticles(u User, namedSQL string, paging ...int) ([]Article, error) {
	var articles []Article

	limit, offset := pagingLimit(paging)

	if err := db.Select(&articles, db.NamedSQL(namedSQL), u.Login, limit, offset); err != nil {
		return articles, err
	}

	return articles, nil
}

func pagingLimit(paging []int) (int, int) {
	limit := 50
	offset := 0

	if len(paging) > 0 {
		limit = paging[0]
		if len(paging) > 1 {
			offset = paging[1]
		}
	}

	return limit, offset
}

func init() {
	sql_stmt["generic:get_feed"] = get_feed
	sql_stmt["generic:create_feed"] = create_feed
	sql_stmt["generic:update_feed"] = update_feed
	sql_stmt["generic:delete_feed"] = delete_feed
	sql_stmt["generic:get_feeds"] = get_feeds
	sql_stmt["generic:get_feed_by_link"] = get_feed_by_link
	sql_stmt["generic:get_unsubscribed_feeds"] = get_unsubscribed_feeds
	sql_stmt["generic:get_user_feed"] = get_user_feed
	sql_stmt["generic:create_user_feed"] = create_user_feed
	sql_stmt["generic:delete_user_feed"] = delete_user_feed
	sql_stmt["generic:get_user_feeds"] = get_user_feeds
	sql_stmt["generic:get_feed_article"] = get_feed_article
	sql_stmt["generic:update_feed_article"] = update_feed_article
	sql_stmt["generic:get_feed_articles"] = get_feed_articles
	sql_stmt["generic:get_feed_articles_desc"] = get_feed_articles_desc
	sql_stmt["generic:get_unread_feed_articles"] = get_unread_feed_articles
	sql_stmt["generic:get_unread_feed_articles_desc"] = get_unread_feed_articles_desc
	sql_stmt["generic:get_read_feed_articles"] = get_read_feed_articles
	sql_stmt["generic:get_user_articles"] = get_user_articles
	sql_stmt["generic:get_user_articles_desc"] = get_user_articles_desc
	sql_stmt["generic:get_unread_user_articles"] = get_unread_user_articles
	sql_stmt["generic:get_unread_user_articles_desc"] = get_unread_user_articles_desc
	sql_stmt["generic:get_read_user_articles"] = get_read_user_articles
	sql_stmt["generic:get_all_articles"] = get_all_articles
	sql_stmt["generic:get_all_feed_articles"] = get_all_feed_articles
	sql_stmt["generic:create_all_user_articles_read_by_date"] = create_all_user_articles_read_by_date
	sql_stmt["generic:delete_all_user_articles_read_by_date"] = delete_all_user_articles_read_by_date
	sql_stmt["generic:create_all_users_articles_read_by_feed_date"] = create_all_users_articles_read_by_feed_date
	sql_stmt["generic:delete_all_users_articles_read_by_feed_date"] = delete_all_users_articles_read_by_feed_date
	sql_stmt["generic:get_user_favorite_articles"] = get_user_favorite_articles
	sql_stmt["generic:get_user_favorite_articles_desc"] = get_user_favorite_articles_desc
}
