package readeef

import "github.com/jmoiron/sqlx"

const (
	exists_feed = `SELECT 1 FROM feeds WHERE link = ?`
	create_feed = `INSERT INTO feeds(title, description, hub_link, link) VALUES(?, ?, ?, ?)`
	update_feed = `UPDATE feeds SET title = ?, description = ?, hub_link = ? WHERE link = ?`
	delete_feed = `DELETE FROM feeds WHERE link = ?`

	exists_user_feed = `SELECT 1 FROM users_feeds WHERE user_login = ? AND feed_link = ?`
	create_user_feed = `INSERT INTO users_feeds(user_login, feed_link) VALUES(?, ?)`
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
	ON uf.feed_link = articles.feed_link AND uf.feed_link = ? AND uf.user_login = ?
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id
LIMIT ?
OFFSET ?
`

	get_user_favorite_articles = `
SELECT uf.feed_link, a.id, a.title, a.description, a.link, a.date
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_link = articles.feed_link AND uf.user_login = ?
INNER JOIN users_articles_fav af
	ON a.id = af.article_id
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id
LIMIT ?
OFFSET ?
`
)

func (db DB) UpdateFeed(f Feed) error {
	if err := f.Validate(); err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	exists := 0
	var stmt *sqlx.Stmt

	row := db.QueryRow(exists_feed, f.Link)
	row.Scan(&exists)

	if exists == 1 {
		stmt, err = tx.Preparex(update_feed)
	} else {
		stmt, err = tx.Preparex(create_feed)
	}

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(f.Title, f.Description, f.HubLink, f.Link)
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
	if err != nil {
		return err
	}

	exists := 0
	var stmt *sqlx.Stmt

	row := db.QueryRow(exists_feed, f.Link)
	row.Scan(&exists)

	if exists != 1 {
		return nil
	}
	stmt, err = tx.Preparex(delete_feed)

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
func (db DB) CreateUserFeed(u User, f Feed) error {
	if err := u.Validate(); err != nil {
		return err
	}
	if err := f.Validate(); err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	exists := 0
	var stmt *sqlx.Stmt

	row := db.QueryRow(exists_user_feed, u.Login, f.Link)
	row.Scan(&exists)

	if exists == 1 {
		return nil
	}

	stmt, err = tx.Preparex(create_user_feed)

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Login, f.Link)
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func (db DB) DeleteUserFeed(u User, f Feed) error {
	if err := u.Validate(); err != nil {
		return err
	}
	if err := f.Validate(); err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	exists := 0
	var stmt *sqlx.Stmt

	row := db.QueryRow(exists_user_feed, u.Login, f.Link)
	row.Scan(&exists)

	if exists != 1 {
		return nil
	}

	stmt, err = tx.Preparex(delete_user_feed)

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Login, f.Link)
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

func (db DB) GetFeedArticles(f Feed, paging ...int) (Feed, error) {
	var articles []Article

	limit, offset := pagingLimit(paging)

	if err := db.Select(&articles, get_feed_articles, f.Link, f.User.Login, limit, offset); err != nil {
		return f, err
	}

	f.Articles = articles

	return f, nil
}

func (db DB) GetUserFavoriteArticles(u User, paging ...int) ([]Article, error) {
	var articles []Article
	limit, offset := pagingLimit(paging)

	if err := db.Select(&articles, get_user_favorite_articles, u.Login, limit, offset); err != nil {
		return articles, err
	}

	return articles, nil
}

func pagingLimit(paging []int) (int, int) {
	limit := 20
	offset := 0

	if len(paging) > 0 {
		limit = paging[0]
		if len(paging) > 1 {
			offset = paging[1]
		}
	}

	return limit, offset
}
