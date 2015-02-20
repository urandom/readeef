package readeef

import "time"

const (
	get_user_tags      = `SELECT DISTINCT tag FROM users_feeds_tags WHERE user_login = $1`
	get_user_feed_tags = `SELECT tag FROM users_feeds_tags WHERE user_login = $1 AND feed_id = $2`

	create_user_feed_tag = `
INSERT INTO users_feeds_tags(user_login, feed_id, tag)
	SELECT $1, $2, $3 EXCEPT SELECT user_login, feed_id, tag
		FROM users_feeds_tags
		WHERE user_login = $1 AND feed_id = $2 AND tag = $3
`

	delete_user_feed_tag = `
DELETE FROM users_feeds_tags WHERE user_login = $1 AND feed_id = $2 AND tag = $3
`

	get_user_tag_feeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds_tags uft
WHERE f.id = uft.feed_id
	AND uft.user_login = $1 AND uft.tag = $2
ORDER BY LOWER(f.title)
`

	get_user_feed_ids_tags = `SELECT feed_id, tag FROM users_feeds_tags WHERE user_login = $1 ORDER BY feed_id`

	create_all_user_tag_articles_read_by_date = `
INSERT INTO users_articles_read
	SELECT uf.user_login, a.id, uf.feed_id
	FROM users_feeds uf INNER JOIN users_feeds_tags uft
		ON uft.feed_id = uf.feed_id AND uft.user_login = uf.user_login
			AND uft.user_login = $1 AND uft.tag = $2
	INNER JOIN articles a
		ON uf.feed_id = a.feed_id
		AND a.id IN (SELECT id FROM articles WHERE date IS NULL OR date < $3)
`

	delete_all_user_tag_articles_read_by_date = `
DELETE FROM users_articles_read WHERE user_login = $1
	AND article_feed_id IN (
		SELECT feed_id FROM users_feeds_tags WHERE user_login = $1 AND tag = $2
	) AND article_id IN (
		SELECT id FROM articles WHERE date IS NULL OR date < $3
	)
`
)

type feedIdTag struct {
	FeedId int64 `db:"feed_id"`
	Tag    string
}

func (db DB) GetUserTags(u User) ([]string, error) {
	var tags []string
	var dbFeed []feedIdTag

	if err := db.Select(&dbFeed, db.NamedSQL("get_user_tags"), u.Login); err != nil {
		return tags, err
	}

	for _, t := range dbFeed {
		tags = append(tags, t.Tag)
	}

	return tags, nil
}

func (db DB) GetUserFeedTags(u User, f Feed) ([]string, error) {
	var tags []string
	var dbFeed []feedIdTag

	if err := db.Select(&dbFeed, db.NamedSQL("get_user_feed_tags"), u.Login, f.Id); err != nil {
		return tags, err
	}

	for _, t := range dbFeed {
		tags = append(tags, t.Tag)
	}

	return tags, nil
}

func (db DB) CreateUserFeedTag(f Feed, tags ...string) error {
	if err := f.Validate(); err != nil {
		return err
	}
	if err := f.User.Validate(); err != nil {
		return err
	}

	// FIXME: Remove when the 'FOREIGN KEY constraing failed' bug is removed
	if db.driver == "sqlite3" {
		db.Query("SELECT 1")
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.NamedSQL("create_user_feed_tag"))

	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, tag := range tags {
		_, err = stmt.Exec(f.User.Login, f.Id, tag)
		if err != nil {
			return err
		}
	}

	tx.Commit()

	return nil
}

func (db DB) DeleteUserFeedTag(f Feed, tags ...string) error {
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

	stmt, err := tx.Preparex(db.NamedSQL("delete_user_feed_tag"))

	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, tag := range tags {
		_, err = stmt.Exec(f.User.Login, f.Id, tag)
		if err != nil {
			return err
		}
	}

	tx.Commit()

	return nil
}

func (db DB) GetUserTagFeeds(u User, tag string) ([]Feed, error) {
	var feeds []Feed

	if err := db.Select(&feeds, db.NamedSQL("get_user_tag_feeds"), u.Login, tag); err != nil {
		return feeds, err
	}

	for _, f := range feeds {
		f.User = u
	}

	return feeds, nil
}

func (db DB) GetUserTagsFeeds(u User) ([]Feed, error) {
	var feedIdTags []feedIdTag

	if err := db.Select(&feedIdTags, db.NamedSQL("get_user_feed_ids_tags"), u.Login); err != nil {
		return []Feed{}, err
	}

	if feeds, err := db.GetUserFeeds(u); err == nil {
		feedMap := make(map[int64]int)

		for i := range feeds {
			feedMap[feeds[i].Id] = i
		}

		for _, tuple := range feedIdTags {
			if i, ok := feedMap[tuple.FeedId]; ok {
				feeds[i].Tags = append(feeds[i].Tags, tuple.Tag)
			}
		}

		return feeds, nil
	} else {
		return []Feed{}, err
	}
}

func (db DB) GetUserTagArticles(u User, tag string, paging ...int) ([]Article, error) {
	return db.getArticles(u, "",
		"INNER JOIN users_feeds_tags uft ON uft.feed_id = uf.feed_id AND uft.user_login = uf.user_login",
		"uft.tag = $2", "read, a.date", []interface{}{tag}, paging...)
}

func (db DB) GetUserTagArticlesDesc(u User, tag string, paging ...int) ([]Article, error) {
	return db.getArticles(u, "",
		"INNER JOIN users_feeds_tags uft ON uft.feed_id = uf.feed_id AND uft.user_login = uf.user_login",
		"uft.tag = $2", "read ASC, a.date DESC", []interface{}{tag}, paging...)
}

func (db DB) GetUnreadUserTagArticles(u User, tag string, paging ...int) ([]Article, error) {
	return db.getArticles(u, "",
		"INNER JOIN users_feeds_tags uft ON uft.feed_id = uf.feed_id AND uft.user_login = uf.user_login",
		"uft.tag = $2 AND ar.article_id IS NULL", "a.date", []interface{}{tag}, paging...)
}

func (db DB) GetUnreadUserTagArticlesDesc(u User, tag string, paging ...int) ([]Article, error) {
	return db.getArticles(u, "",
		"INNER JOIN users_feeds_tags uft ON uft.feed_id = uf.feed_id AND uft.user_login = uf.user_login",
		"uft.tag = $2 AND ar.article_id IS NULL", "a.date DESC", []interface{}{tag}, paging...)
}

func (db DB) MarkUserTagArticlesByDateAsRead(u User, tag string, d time.Time, read bool) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.NamedSQL("delete_all_user_tag_articles_read_by_date"))

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Login, tag, d)
	if err != nil {
		return err
	}

	stmt, err = tx.Preparex(db.NamedSQL("create_all_user_tag_articles_read_by_date"))

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Login, tag, d)
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func init() {
	sqlStmt["generic:get_user_tags"] = get_user_tags
	sqlStmt["generic:get_user_feed_tags"] = get_user_feed_tags
	sqlStmt["generic:create_user_feed_tag"] = create_user_feed_tag
	sqlStmt["generic:delete_user_feed_tag"] = delete_user_feed_tag
	sqlStmt["generic:get_user_tag_feeds"] = get_user_tag_feeds
	sqlStmt["generic:get_user_feed_ids_tags"] = get_user_feed_ids_tags
	sqlStmt["generic:create_all_user_tag_articles_read_by_date"] = create_all_user_tag_articles_read_by_date
	sqlStmt["generic:delete_all_user_tag_articles_read_by_date"] = delete_all_user_tag_articles_read_by_date
}
