package readeef

const (
	get_user_tags      = `SELECT tag FROM users_feeds_tags WHERE user_login = $1`
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

	get_user_tag_articles = `
SELECT uf.feed_id, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
INNER JOIN users_feeds_tags uft
	ON uft.feed_id = uf.feed_id AND uft.user_login = uf.user_login AND uft.tag = $2
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_id = ar.article_feed_id
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_id = af.article_feed_id
ORDER BY a.date
LIMIT $3
OFFSET $4
`

	get_user_tag_articles_desc = `
SELECT uf.feed_id, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
INNER JOIN users_feeds_tags uft
	ON uft.feed_id = uf.feed_id AND uft.user_login = uf.user_login AND uft.tag = $2
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_id = ar.article_feed_id
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_id = af.article_feed_id
ORDER BY a.date DESC
LIMIT $3
OFFSET $4
`

	get_unread_user_tag_articles = `
SELECT uf.feed_id, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
INNER JOIN users_feeds_tags uft
	ON uft.feed_id = uf.feed_id AND uft.user_login = uf.user_login AND uft.tag = $2
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_id = ar.article_feed_id
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_id = af.article_feed_id
WHERE ar.article_id IS NULL
ORDER BY a.date
LIMIT $3
OFFSET $4
`

	get_unread_user_tag_articles_desc = `
SELECT uf.feed_id, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
INNER JOIN users_feeds_tags uft
	ON uft.feed_id = uf.feed_id AND uft.user_login = uf.user_login AND uft.tag = $2
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND a.feed_id = ar.article_feed_id
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND a.feed_id = af.article_feed_id
WHERE ar.article_id IS NULL
ORDER BY a.date DESC
LIMIT $3
OFFSET $4
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

		for i := 0; i < len(feeds); i++ {
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
	return db.getUserTagArticles(u, tag, "get_user_tag_articles", paging...)
}

func (db DB) GetUserTagArticlesDesc(u User, tag string, paging ...int) ([]Article, error) {
	return db.getUserTagArticles(u, tag, "get_user_tag_articles_desc", paging...)
}

func (db DB) GetUnreadUserTagArticles(u User, tag string, paging ...int) ([]Article, error) {
	return db.getUserTagArticles(u, tag, "get_unread_user_tag_articles", paging...)
}

func (db DB) GetUnreadUserTagArticlesDesc(u User, tag string, paging ...int) ([]Article, error) {
	return db.getUserTagArticles(u, tag, "get_unread_user_tag_articles_desc", paging...)
}

func (db DB) getUserTagArticles(u User, tag, namedSQL string, paging ...int) ([]Article, error) {
	var articles []Article

	limit, offset := pagingLimit(paging)

	if err := db.Select(&articles, db.NamedSQL(namedSQL), u.Login, tag, limit, offset); err != nil {
		return articles, err
	}

	return articles, nil
}

func init() {
	sql_stmt["generic:get_user_tags"] = get_user_tags
	sql_stmt["generic:get_user_feed_tags"] = get_user_feed_tags
	sql_stmt["generic:create_user_feed_tag"] = create_user_feed_tag
	sql_stmt["generic:delete_user_feed_tag"] = delete_user_feed_tag
	sql_stmt["generic:get_user_tag_feeds"] = get_user_tag_feeds
	sql_stmt["generic:get_user_feed_ids_tags"] = get_user_feed_ids_tags
	sql_stmt["generic:get_user_tag_articles"] = get_user_tag_articles
	sql_stmt["generic:get_user_tag_articles_desc"] = get_user_tag_articles_desc
	sql_stmt["generic:get_unread_user_tag_articles"] = get_unread_user_tag_articles
	sql_stmt["generic:get_unread_user_tag_articles_desc"] = get_unread_user_tag_articles_desc
}
