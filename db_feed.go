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

	update_feed_article = `
UPDATE articles SET title = $1, description = $2, date = $3 WHERE feed_id = $4 AND link = $5
`

	update_feed_article_with_guid = `
UPDATE articles SET title = $1, description = $2, date = $3 WHERE feed_id = $4 AND guid = $5
`

	get_article_columns = `
SELECT a.feed_id, a.id, a.title, a.description, a.link, a.date, a.guid,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
`

	get_article_tables = `
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id
`

	get_article_joins = `
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND uf.user_login = ar.user_login
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND uf.user_login = af.user_login
`

	get_user_article_count = `
SELECT count(a.id)
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
`

	get_all_articles = `
SELECT a.feed_id, a.id, a.title, a.description, a.link, a.guid, a.date
FROM articles a
`

	get_all_feed_articles = `
SELECT a.feed_id, a.id, a.title, a.description, a.link, a.guid, a.date
FROM articles a
WHERE a.feed_id = $1
`

	get_all_unread_user_article_ids = `
SELECT a.id
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND uf.user_login = ar.user_login
WHERE ar.article_id IS NULL
`

	get_all_favorite_user_article_ids = `
SELECT a.id
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND uf.user_login = af.user_login
WHERE af.article_id IS NOT NULL
`

	create_all_user_articles_read_by_date = `
INSERT INTO users_articles_read
	SELECT uf.user_login, a.id
	FROM users_feeds uf INNER JOIN articles a
		ON uf.feed_id = a.feed_id AND uf.user_login = $1
		AND a.id IN (SELECT id FROM articles WHERE date IS NULL OR date < $2)
`

	delete_all_user_articles_read_by_date = `
DELETE FROM users_articles_read WHERE user_login = $1 AND article_id IN (
	SELECT id FROM articles WHERE date IS NULL OR date < $2
)
`

	delete_newer_user_articles_read_by_date = `
DELETE FROM users_articles_read WHERE user_login = $1 AND article_id IN (
	SELECT id FROM articles WHERE date > $2
)
`

	create_all_users_articles_read_by_feed_date = `
INSERT INTO users_articles_read
	SELECT uf.user_login, a.id
	FROM users_feeds uf INNER JOIN articles a
		ON uf.feed_id = a.feed_id AND uf.user_login = $1 AND uf.feed_id = $2
		AND a.id IN (SELECT id FROM articles WHERE date IS NULL OR date < $3)
`

	delete_all_users_articles_read_by_feed_date = `
DELETE FROM users_articles_read WHERE user_login = $1 AND article_id IN (
	SELECT id FROM articles WHERE feed_id = $2 AND (date IS NULL OR date < $3)
)
`
)

var (
	ErrNoFeedUser = errors.New("Feed does not have an associated user.")
)

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

		for i := range f.Articles {
			f.Articles[i].FeedId = id
		}
	}

	f.lastUpdatedArticleLinks = map[string]bool{}
	newArticles, err := db.updateFeedArticles(tx, f, f.Articles)
	if err != nil {
		return f, false, err
	}

	tx.Commit()

	for _, a := range newArticles {
		f.lastUpdatedArticleLinks[a.Link] = true
	}

	if articles, err := db.populateArticleIds(f.Articles); err == nil {
		f.Articles = articles
	} else {
		return f, false, err
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

func (db DB) GetFeedArticle(articleId int64, user User) (Article, error) {
	Debug.Printf("Getting feed article %d\n", articleId)

	articles, err := db.getArticles(user, "", "", "a.id = $2", "", []interface{}{articleId})
	if err == nil {
		return articles[0], nil
	} else {
		return Article{}, err
	}
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

	tx.Commit()

	if articles, err := db.populateArticleIds(articles); err == nil {
		f.Articles = append(f.Articles, articles...)
	} else {
		return f, err
	}

	return f, nil
}

func (db DB) GetFeedArticles(f Feed, paging ...int) (Feed, error) {
	return db.getFeedArticles(f, "", "read, a.date", paging...)
}

func (db DB) GetFeedArticlesDesc(f Feed, paging ...int) (Feed, error) {
	return db.getFeedArticles(f, "", "read ASC, a.date DESC", paging...)
}

func (db DB) GetUnreadFeedArticles(f Feed, paging ...int) (Feed, error) {
	return db.getFeedArticles(f, "ar.article_id IS NULL", "a.date", paging...)
}

func (db DB) GetUnreadFeedArticlesDesc(f Feed, paging ...int) (Feed, error) {
	return db.getFeedArticles(f, "ar.article_id IS NULL", "a.date DESC", paging...)
}

func (db DB) GetReadFeedArticles(f Feed, paging ...int) (Feed, error) {
	return db.getFeedArticles(f, "ar.article_id IS NOT NULL", "a.date", paging...)
}

func (db DB) GetUserArticleCount(u User) (int64, error) {
	var count int64 = -1

	Debug.Println("Getting user article count")

	if err := db.Get(&count, db.NamedSQL("get_user_article_count"), u.Login); err != nil {
		return count, err
	}

	return count, nil
}

func (db DB) GetUnorderedUserArticles(u User, since int64, paging ...int) ([]Article, error) {
	return db.getArticles(u, "", "", "a.id > $2", "a.id", []interface{}{since}, paging...)
}

func (db DB) GetUnorderedUserArticlesDesc(u User, max int64, paging ...int) ([]Article, error) {
	return db.getArticles(u, "", "", "a.id < $2", "a.id DESC", []interface{}{max}, paging...)
}

func (db DB) GetUserArticles(u User, paging ...int) ([]Article, error) {
	return db.getArticles(u, "", "", "", "read, a.date", nil, paging...)
}

func (db DB) GetUserArticlesDesc(u User, paging ...int) ([]Article, error) {
	return db.getArticles(u, "", "", "", "read ASC, a.date DESC", nil, paging...)
}

func (db DB) GetUnreadUserArticles(u User, paging ...int) ([]Article, error) {
	return db.getArticles(u, "", "", "ar.article_id IS NULL", "a.date", nil, paging...)
}

func (db DB) GetUnreadUserArticlesDesc(u User, paging ...int) ([]Article, error) {
	return db.getArticles(u, "", "", "ar.article_id IS NULL", "a.date DESC", nil, paging...)
}

func (db DB) GetReadUserArticles(u User, paging ...int) ([]Article, error) {
	return db.getArticles(u, "", "", "ar.article_id IS NOT NULL", "a.date", nil, paging...)
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

func (db DB) GetSpecificUserArticles(u User, ids ...int64) ([]Article, error) {
	if len(ids) == 0 {
		return []Article{}, nil
	}

	where := "("

	args := []interface{}{}
	index := 1
	for _, id := range ids {
		if index > 1 {
			where += ` OR `
		}

		where += fmt.Sprintf(`a.id = $%d`, index+1)
		args = append(args, id)
		index = len(args) + 1
	}

	where += ")"

	return db.getArticles(u, "", "", where, "", args)
}

func (db DB) GetAllUnreadUserArticleIds(u User) ([]int64, error) {
	var ids []int64

	if err := db.Select(&ids, db.NamedSQL("get_all_unread_user_article_ids"), u.Login); err != nil {
		return ids, err
	}

	return ids, nil
}

func (db DB) GetAllFavoriteUserArticleIds(u User) ([]int64, error) {
	var ids []int64

	if err := db.Select(&ids, db.NamedSQL("get_all_favorite_user_article_ids"), u.Login); err != nil {
		return ids, err
	}

	return ids, nil
}

func (db DB) MarkUserArticlesAsRead(u User, articles []Article, read bool) error {
	if len(articles) == 0 {
		return nil
	}

	var sql string
	var args []interface{}

	if read {
		sql = `INSERT INTO users_articles_read(user_login, article_id) `
	} else {
		sql = `DELETE FROM users_articles_read WHERE `
	}

	index := 1
	for _, a := range articles {
		if err := a.Validate(); err != nil {
			return err
		}

		if a.Id == 0 {
			return ValidationError{errors.New("Article has no id")}
		}

		if read {
			if index > 1 {
				sql += ` UNION `
			}

			sql += fmt.Sprintf(`SELECT $%d, $%d EXCEPT SELECT user_login, article_id FROM users_articles_read WHERE user_login = $%d AND article_id = $%d`, index, index+1, index, index+1)
			args = append(args, u.Login, a.Id)
		} else {
			if index > 1 {
				sql += `OR `
			}

			sql += fmt.Sprintf(`(user_login = $%d AND article_id = $%d)`, index, index+1)
			args = append(args, u.Login, a.Id)
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

func (db DB) MarkNewerUserArticlesByDateAsUnread(u User, d time.Time, read bool) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.NamedSQL("delete_newer_user_articles_read_by_date"))

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Login, d)
	if err != nil {
		return err
	}

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
	return db.getArticles(u, "", "", "af.article_id IS NOT NULL", "a.date", nil, paging...)
}

func (db DB) GetUserFavoriteArticlesDesc(u User, paging ...int) ([]Article, error) {
	return db.getArticles(u, "", "", "af.article_id IS NOT NULL", "a.date DESC", nil, paging...)
}

func (db DB) MarkUserArticlesAsFavorite(u User, articles []Article, read bool) error {
	if len(articles) == 0 {
		return nil
	}

	var sql string
	var args []interface{}

	if read {
		sql = `INSERT INTO users_articles_fav(user_login, article_id) `
	} else {
		sql = `DELETE FROM users_articles_fav WHERE `
	}

	index := 1
	for _, a := range articles {
		if err := a.Validate(); err != nil {
			return err
		}

		if a.Id == 0 {
			return ValidationError{errors.New("Article has no id")}
		}

		if read {
			if index > 1 {
				sql += ` UNION `
			}

			sql += fmt.Sprintf(`SELECT $%d, $%d EXCEPT SELECT user_login, article_id FROM users_articles_fav WHERE user_login = $%d AND article_id = $%d`, index, index+1, index, index+1)
			args = append(args, u.Login, a.Id)
		} else {
			if index > 1 {
				sql += `OR `
			}

			sql += fmt.Sprintf(`(user_login = $%d AND article_id = $%d)`, index, index+1)
			args = append(args, u.Login, a.Id)
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

		var sql string
		args := []interface{}{a.Title, a.Description, a.Date, f.Id}

		if a.Guid.Valid {
			sql = db.NamedSQL("update_feed_article_with_guid")
			args = append(args, a.Guid)
		} else {
			sql = db.NamedSQL("update_feed_article")
			args = append(args, a.Link)
		}

		stmt, err := tx.Preparex(sql)
		if err != nil {
			return newArticles, err
		}
		defer stmt.Close()

		res, err := stmt.Exec(args...)
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

	sql := `INSERT INTO articles(feed_id, link, guid, title, description, date) `
	args := []interface{}{}
	index := 1

	for _, a := range newArticles {
		if err := a.Validate(); err != nil {
			return newArticles, err
		}

		if index > 1 {
			sql += ` UNION `
		}

		sql += fmt.Sprintf(
			`SELECT $%d, $%d, $%d ,$%d, $%d, $%d EXCEPT SELECT feed_id, link, guid, title, description, date FROM articles WHERE feed_id = $%d AND link = $%d`,
			index, index+1, index+2, index+3, index+4, index+5, index, index+1)
		args = append(args, f.Id, a.Link, a.Guid, a.Title, a.Description, a.Date)
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

func (db DB) getFeedArticles(f Feed, where, order string, paging ...int) (Feed, error) {
	if f.User.Login == "" {
		return f, ErrNoFeedUser
	}

	var articles []Article

	if where == "" {
		where = "uf.feed_id = $2"
	} else {
		where = "uf.feed_id = $2 AND " + where
	}

	articles, err := db.getArticles(f.User, "", "", where, order, []interface{}{f.Id}, paging...)
	if err != nil {
		return f, err
	}

	f.Articles = articles

	return f, nil
}

func (db DB) getArticles(u User, columns, join, where, order string, args []interface{}, paging ...int) ([]Article, error) {
	var articles []Article

	sql := db.NamedSQL("get_article_columns")
	if columns != "" {
		sql += ", " + columns
	}

	sql += db.NamedSQL("get_article_tables")
	if join != "" {
		sql += " " + join
	}

	sql += db.NamedSQL("get_article_joins")

	args = append([]interface{}{u.Login}, args...)
	sql += " WHERE uf.user_login = $1"
	if where != "" {
		sql += " AND " + where
	}

	if order != "" {
		sql += " ORDER BY " + order
	}

	if len(paging) > 0 {
		limit, offset := pagingLimit(paging)

		sql += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
		args = append(args, limit, offset)
	}

	if err := db.Select(&articles, sql, args...); err != nil {
		return articles, err
	}

	return articles, nil
}

func (db DB) populateArticleIds(articles []Article) ([]Article, error) {
	if len(articles) == 0 {
		return articles, nil
	}

	temp := []Article{}
	sql := `
SELECT a.feed_id, a.id, a.link
FROM articles a
WHERE `

	args := []interface{}{}
	index := 1

	for _, a := range articles {
		if err := a.Validate(); err != nil {
			return articles, err
		}

		if a.Id != 0 {
			continue
		}

		if index > 1 {
			sql += ` OR `
		}

		sql += fmt.Sprintf(`(feed_id = $%d AND link = $%d)`, index, index+1)
		args = append(args, a.FeedId, a.Link)
		index = len(args) + 1
	}

	if index == 1 {
		return articles, nil
	}

	if err := db.Select(&temp, sql, args...); err != nil {
		return articles, err
	}

	articleMap := map[string]int{}
	for i, a := range articles {
		articleMap[fmt.Sprintf("%d:%s", a.FeedId, a.Link)] = i
	}

	for _, a := range temp {
		if i, ok := articleMap[fmt.Sprintf("%d:%s", a.FeedId, a.Link)]; ok {
			articles[i].Id = a.Id
		}
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
	sql_stmt["generic:update_feed_article"] = update_feed_article
	sql_stmt["generic:update_feed_article_with_guid"] = update_feed_article_with_guid
	sql_stmt["generic:get_article_columns"] = get_article_columns
	sql_stmt["generic:get_article_tables"] = get_article_tables
	sql_stmt["generic:get_article_joins"] = get_article_joins
	sql_stmt["generic:get_user_article_count"] = get_user_article_count
	sql_stmt["generic:get_all_articles"] = get_all_articles
	sql_stmt["generic:get_all_feed_articles"] = get_all_feed_articles
	sql_stmt["generic:get_all_unread_user_article_ids"] = get_all_unread_user_article_ids
	sql_stmt["generic:get_all_favorite_user_article_ids"] = get_all_favorite_user_article_ids
	sql_stmt["generic:create_all_user_articles_read_by_date"] = create_all_user_articles_read_by_date
	sql_stmt["generic:delete_all_user_articles_read_by_date"] = delete_all_user_articles_read_by_date
	sql_stmt["generic:delete_newer_user_articles_read_by_date"] = delete_newer_user_articles_read_by_date
	sql_stmt["generic:create_all_users_articles_read_by_feed_date"] = create_all_users_articles_read_by_feed_date
	sql_stmt["generic:delete_all_users_articles_read_by_feed_date"] = delete_all_users_articles_read_by_feed_date
}
