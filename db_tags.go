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
)

type feedIdTag struct {
	FeedId int64 `db:"feed_id"`
	Tag    string
}

func (db DB) GetUserTags(u User) ([]string, error) {
	var tags []string

	if err := db.Select(&tags, db.NamedSQL("get_user_tags"), u.Login); err != nil {
		return tags, err
	}

	return tags, nil
}

func (db DB) GetUserFeedTags(u User, f Feed) ([]string, error) {
	var tags []string

	if err := db.Select(&tags, db.NamedSQL("get_user_feed_tags"), u.Login, f.Id); err != nil {
		return tags, err
	}

	return tags, nil
}

func (db DB) CreateUserFeedTag(f Feed, tag string) error {
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

	_, err = stmt.Exec(f.User.Login, f.Id, tag)
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func (db DB) DeleteUserFeedTag(f Feed, tag string) error {
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

	_, err = stmt.Exec(f.User.Login, f.Id, tag)
	if err != nil {
		return err
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

func init() {
	sql_stmt["generic:get_user_tags"] = get_user_tags
	sql_stmt["generic:get_user_feed_tags"] = get_user_feed_tags
	sql_stmt["generic:create_user_feed_tag"] = create_user_feed_tag
	sql_stmt["generic:delete_user_feed_tag"] = delete_user_feed_tag
	sql_stmt["generic:get_user_tag_feeds"] = get_user_tag_feeds
	sql_stmt["generic:get_user_feed_ids_tags"] = get_user_feed_ids_tags
}
