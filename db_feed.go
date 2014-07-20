package readeef

import (
	"errors"
	"time"
)

const (
	get_feed    = `SELECT title, description, hub_link FROM feeds WHERE link = ?`
	create_feed = `
INSERT INTO feeds(link, title, description, hub_link)
	SELECT ?, ?, ?, ? EXCEPT SELECT link, title, description, hub_link FROM feeds WHERE link = ?`
	update_feed = `UPDATE feeds SET title = ?, description = ?, hub_link = ? WHERE link = ?`
	delete_feed = `DELETE FROM feeds WHERE link = ?`

	get_feeds = `SELECT link, title, description, hub_link FROM feeds`

	create_user_feed = `
INSERT INTO users_feeds(user_login, feed_link)
	SELECT ?, ? EXCEPT SELECT user_login, feed_link FROM users_feeds WHERE user_login = ? AND feed_link = ?`
	delete_user_feed = `DELETE FROM users_feeds WHERE user_login = ? AND feed_link = ?`

	get_user_feeds = `
SELECT f.link, f.title, f.description, f.link, f.hub_link
FROM feeds f, users_feeds uf
WHERE f.link = uf.feed_link
	AND uf.user_login = ?
`

	get_feed_articles = `
SELECT uf.feed_link, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_link = a.feed_link AND uf.feed_link = ? AND uf.user_login = ?
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_link = ar.article_feed_link
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_link = af.article_feed_link
ORDER BY a.date
LIMIT ?
OFFSET ?
`

	get_unread_feed_articles = `
SELECT uf.feed_link, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_link = a.feed_link AND uf.feed_link = ? AND uf.user_login = ?
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_link = ar.article_feed_link
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_link = af.article_feed_link
WHERE ar.article_id IS NULL
ORDER BY a.date
LIMIT ?
OFFSET ?
`

	get_read_feed_articles = `
SELECT uf.feed_link, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_link = a.feed_link AND uf.feed_link = ? AND uf.user_login = ?
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_link = ar.article_feed_link
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_link = af.article_feed_link
WHERE ar.article_id IS NOT NULL
ORDER BY a.date
LIMIT ?
OFFSET ?
`

	get_user_articles = `
SELECT uf.feed_link, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_link = a.feed_link AND uf.user_login = ?
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_link = ar.article_feed_link
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_link = af.article_feed_link
ORDER BY a.date
LIMIT ?
OFFSET ?
`

	get_unread_user_articles = `
SELECT uf.feed_link, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_link = a.feed_link AND uf.user_login = ?
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_link = ar.article_feed_link
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_link = af.article_feed_link
WHERE ar.article_id IS NULL
ORDER BY a.date
LIMIT ?
OFFSET ?
`

	get_read_user_articles = `
SELECT uf.feed_link, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_link = a.feed_link AND uf.user_login = ?
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_link = ar.article_feed_link
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_link = af.article_feed_link
WHERE ar.article_id IS NOT NULL
ORDER BY a.date
LIMIT ?
OFFSET ?
`

	create_all_users_articles_read_by_date = `
INSERT INTO users_articles_read
	SELECT uf.user_login, a.id, uf.feed_link
	FROM users_feeds uf INNER JOIN articles a
		ON uf.feed_link = a.feed_link AND uf.user_login = ?
		AND a.id IN (SELECT id FROM articles WHERE date IS NULL OR date < ?)
`

	delete_all_users_articles_read_by_date = `
DELETE FROM users_articles_read WHERE user_login = ? AND article_id IN (
	SELECT id FROM articles WHERE date IS NULL OR date < ?
)
`

	create_all_users_articles_read_by_feed_date = `
INSERT INTO users_articles_read
	SELECT uf.user_login, a.id, uf.feed_link
	FROM users_feeds uf INNER JOIN articles a
		ON uf.feed_link = a.feed_link AND uf.user_login = ? AND uf.feed_link = ?
		AND a.id IN (SELECT id FROM articles WHERE date IS NULL OR date < ?)
`

	delete_all_users_articles_read_by_feed_date = `
DELETE FROM users_articles_read WHERE user_login = ? AND article_feed_link = ? AND article_id IN (
	SELECT id FROM articles WHERE date IS NULL OR date < ?
)
`

	get_user_favorite_articles = `
SELECT uf.feed_link, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_link = a.feed_link AND uf.user_login = ?
INNER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_link = af.article_feed_link
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_link = ar.article_feed_link
ORDER BY a.date
LIMIT ?
OFFSET ?
`
)

var (
	ErrNoFeedUser = errors.New("Feed does not have an associated user.")
)

func (db DB) GetFeed(link string) (Feed, error) {
	var f Feed
	if err := db.Get(&f, get_feed, link); err != nil {
		return f, err
	}

	f.Link = link

	return f, nil
}

func (db DB) UpdateFeed(f Feed) error {
	if err := f.Validate(); err != nil {
		return err
	}

	tx, err := db.Beginx()
	defer tx.Rollback()
	if err != nil {
		return err
	}

	ustmt, err := tx.Preparex(update_feed)
	if err != nil {
		return err
	}
	defer ustmt.Close()

	_, err = ustmt.Exec(f.Title, f.Description, f.HubLink, f.Link)
	if err != nil {
		return err
	}

	cstmt, err := tx.Preparex(create_feed)

	if err != nil {
		return err
	}
	defer cstmt.Close()

	_, err = cstmt.Exec(f.Link, f.Title, f.Description, f.HubLink, f.Link)
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func (db DB) DeleteFeed(f Feed) error {
	if err := f.Validate(); err != nil {
		return err
	}

	tx, err := db.Beginx()
	defer tx.Rollback()
	if err != nil {
		return err
	}

	stmt, err := tx.Preparex(delete_feed)

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(f.Link)
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func (db DB) GetFeeds() ([]Feed, error) {
	var feeds []Feed

	if err := db.Select(&feeds, get_feeds); err != nil {
		return feeds, err
	}

	return feeds, nil
}

func (db DB) CreateUserFeed(u User, f Feed) (Feed, error) {
	if err := u.Validate(); err != nil {
		return f, err
	}
	if err := f.Validate(); err != nil {
		return f, err
	}

	tx, err := db.Beginx()
	defer tx.Rollback()
	if err != nil {
		return f, err
	}

	stmt, err := tx.Preparex(create_user_feed)

	if err != nil {
		return f, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Login, f.Link, u.Login, f.Link)
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
	defer tx.Rollback()
	if err != nil {
		return err
	}

	stmt, err := tx.Preparex(delete_user_feed)

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(f.User.Login, f.Link)
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func (db DB) GetUserFeeds(u User) ([]Feed, error) {
	var feeds []Feed

	if err := db.Select(&feeds, get_user_feeds, u.Login); err != nil {
		return feeds, err
	}

	for _, f := range feeds {
		f.User = u
	}

	return feeds, nil
}

func (db DB) CreateFeedArticles(f Feed, articles []Article) (Feed, error) {
	if len(articles) == 0 {
		return f, nil
	}

	sql := `INSERT INTO articles(id, feed_link, title, description, link, date) `
	args := []interface{}{}

	for i, a := range articles {
		if err := a.Validate(); err != nil {
			return f, err
		}

		if i != 0 {
			sql += `UNION `
		}

		sql += `SELECT ?, ?, ?, ? ,?, ? EXCEPT SELECT id, feed_link, title, description, link, date FROM articles WHERE id = ? AND feed_link = ? `
		args = append(args, a.Id, f.Link, a.Title, a.Description, a.Link, a.Date, a.Id, f.Link)
		a.FeedLink = f.Link
	}

	tx, err := db.Beginx()
	defer tx.Rollback()
	if err != nil {
		return f, err
	}

	stmt, err := tx.Preparex(sql)

	if err != nil {
		return f, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return f, err
	}

	f.Articles = append(f.Articles, articles...)

	tx.Commit()

	return f, nil
}

func (db DB) GetFeedArticles(f Feed, paging ...int) (Feed, error) {
	if f.User.Login == "" {
		return f, ErrNoFeedUser
	}

	var articles []Article

	offset, limit := pagingLimit(paging)

	if err := db.Select(&articles, get_feed_articles, f.Link, f.User.Login, limit, offset); err != nil {
		return f, err
	}

	f.Articles = articles

	return f, nil
}

func (db DB) GetUnreadFeedArticles(f Feed, paging ...int) (Feed, error) {
	if f.User.Login == "" {
		return f, ErrNoFeedUser
	}

	var articles []Article

	offset, limit := pagingLimit(paging)

	if err := db.Select(&articles, get_unread_feed_articles, f.Link, f.User.Login, limit, offset); err != nil {
		return f, err
	}

	f.Articles = articles

	return f, nil
}

func (db DB) GetReadFeedArticles(f Feed, paging ...int) (Feed, error) {
	if f.User.Login == "" {
		return f, ErrNoFeedUser
	}

	var articles []Article

	offset, limit := pagingLimit(paging)

	if err := db.Select(&articles, get_read_feed_articles, f.Link, f.User.Login, limit, offset); err != nil {
		return f, err
	}

	f.Articles = articles

	return f, nil
}

func (db DB) GetUserArticles(u User, paging ...int) ([]Article, error) {
	var articles []Article

	offset, limit := pagingLimit(paging)

	if err := db.Select(&articles, get_user_articles, u.Login, limit, offset); err != nil {
		return articles, err
	}

	return articles, nil
}

func (db DB) GetUnreadUserArticles(u User, paging ...int) ([]Article, error) {
	var articles []Article

	offset, limit := pagingLimit(paging)

	if err := db.Select(&articles, get_unread_user_articles, u.Login, limit, offset); err != nil {
		return articles, err
	}

	return articles, nil
}

func (db DB) GetReadUserArticles(u User, paging ...int) ([]Article, error) {
	var articles []Article

	offset, limit := pagingLimit(paging)

	if err := db.Select(&articles, get_read_user_articles, u.Login, limit, offset); err != nil {
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
		sql = `INSERT INTO users_articles_read(user_login, article_id, article_feed_link) `
	} else {
		sql = `DELETE FROM users_articles_read WHERE `
	}
	for i, a := range articles {
		if err := a.Validate(); err != nil {
			return err
		}

		if read {
			if i != 0 {
				sql += `UNION `
			}

			sql += `SELECT ?, ?, ? EXCEPT SELECT user_login, article_id, article_feed_link FROM users_articles_read WHERE user_login = ? AND article_id = ? AND article_feed_link = ?`
			args = append(args, u.Login, a.Id, a.FeedLink, u.Login, a.Id, a.FeedLink)
		} else {
			if i != 0 {
				sql += `OR `
			}

			sql += `(user_login = ? AND article_id = ? AND article_feed_link = ?)`
			args = append(args, u.Login, a.Id, a.FeedLink)
		}
	}

	tx, err := db.Beginx()
	defer tx.Rollback()
	if err != nil {
		return err
	}

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
	defer tx.Rollback()
	if err != nil {
		return err
	}

	stmt, err := tx.Preparex(delete_all_users_articles_read_by_date)

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Login, d)
	if err != nil {
		return err
	}

	stmt, err = tx.Preparex(create_all_users_articles_read_by_date)

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
	defer tx.Rollback()
	if err != nil {
		return err
	}

	stmt, err := tx.Preparex(delete_all_users_articles_read_by_feed_date)

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(f.User.Login, f.Link, d)
	if err != nil {
		return err
	}

	stmt, err = tx.Preparex(create_all_users_articles_read_by_feed_date)

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(f.User.Login, f.Link, d)
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}
func (db DB) GetUserFavoriteArticles(u User, paging ...int) ([]Article, error) {
	var articles []Article

	offset, limit := pagingLimit(paging)

	if err := db.Select(&articles, get_user_favorite_articles, u.Login, limit, offset); err != nil {
		return articles, err
	}

	return articles, nil
}

func (db DB) MarkUserArticlesAsFavorite(u User, articles []Article, read bool) error {
	if len(articles) == 0 {
		return nil
	}

	var sql string
	var args []interface{}

	if read {
		sql = `INSERT INTO users_articles_fav(user_login, article_id, article_feed_link) `
	} else {
		sql = `DELETE FROM users_articles_fav WHERE `
	}
	for i, a := range articles {
		if err := a.Validate(); err != nil {
			return err
		}

		if read {
			if i != 0 {
				sql += ` UNION `
			}

			sql += `SELECT ?, ?, ? EXCEPT SELECT user_login, article_id, article_feed_link FROM users_articles_fav WHERE user_login = ? AND article_id = ? AND article_feed_link = ?`
			args = append(args, u.Login, a.Id, a.FeedLink, u.Login, a.Id, a.FeedLink)
		} else {
			if i != 0 {
				sql += `OR `
			}

			sql += `(user_login = ? AND article_id = ? AND article_feed_link = ?)`
			args = append(args, u.Login, a.Id, a.FeedLink)
		}
	}

	tx, err := db.Beginx()
	defer tx.Rollback()
	if err != nil {
		return err
	}

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

func pagingLimit(paging []int) (int, int) {
	offset := 0
	limit := 50

	if len(paging) > 0 {
		offset = paging[0]
		if len(paging) > 1 {
			limit = paging[1]
		}
	}

	return offset, limit
}
