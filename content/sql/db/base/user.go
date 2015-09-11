package base

func init() {
	sql["create_user"] = createUser
	sql["update_user"] = updateUser
	sql["delete_user"] = deleteUser
	sql["get_user_feed"] = getUserFeed
	sql["create_user_feed"] = createUserFeed
	sql["get_user_feeds"] = getUserFeeds
	sql["get_user_feed_ids_tags"] = getUserFeedIdsTags
	sql["get_user_tags"] = getUserTags
	sql["get_article_columns"] = getArticleColumns
	sql["get_article_tables"] = getArticleTables
	sql["get_article_joins"] = getArticleJoins
	sql["get_all_unread_user_article_ids"] = getAllUnreadUserArticleIds
	sql["get_all_favorite_user_article_ids"] = getAllFavoriteUserArticleIds
	sql["get_user_article_count"] = getUserArticleCount
	sql["create_missing_user_article_state_by_date"] = createMissingUserArticleStateByDate
	sql["update_all_user_article_state_by_date"] = updateAllUserArticleReadStateByDate
	sql["create_newer_missing_user_article_state_by_date"] = createNewerMissingUserArticleStateByDate
	sql["update_all_newer_user_article_state_by_date"] = updateAllNewerUserArticleReadStateByDate
	sql["get_user_unread_count"] = getUserUnreadCount
}

const (
	createUser = `
INSERT INTO users(login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api)
	SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11 EXCEPT
	SELECT login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api FROM users WHERE login = $1`
	updateUser = `
UPDATE users SET first_name = $1, last_name = $2, email = $3, admin = $4, active = $5, profile_data = $6, hash_type = $7, salt = $8, hash = $9, md5_api = $10
	WHERE login = $11`
	deleteUser  = `DELETE FROM users WHERE login = $1`
	getUserFeed = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds uf
WHERE f.id = uf.feed_id
	AND f.id = $1 AND uf.user_login = $2
`
	createUserFeed = `
INSERT INTO users_feeds(user_login, feed_id)
	SELECT $1, $2 EXCEPT SELECT user_login, feed_id FROM users_feeds WHERE user_login = $1 AND feed_id = $2`
	getUserFeeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds uf
WHERE f.id = uf.feed_id
	AND uf.user_login = $1
ORDER BY LOWER(f.title)
`
	getUserFeedIdsTags = `SELECT feed_id, tag FROM users_feeds_tags WHERE user_login = $1 ORDER BY feed_id`
	getUserTags        = `SELECT DISTINCT tag FROM users_feeds_tags WHERE user_login = $1`
	getArticleColumns  = `
SELECT a.feed_id, a.id, a.title, a.description, a.link, a.date, a.guid,
CASE WHEN uas.article_id IS NULL OR NOT uas.read THEN 0 ELSE 1 END AS read,
CASE WHEN uas.article_id IS NULL OR NOT uas.favorite THEN 0 ELSE 1 END AS favorite,
COALESCE(at.thumbnail, '') as thumbnail,
COALESCE(at.link, '') as thumbnail_link
`

	getArticleTables = `
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id
`

	getArticleJoins = `
LEFT OUTER JOIN users_articles_states uas
	ON a.id = uas.article_id AND uf.user_login = uas.user_login
LEFT OUTER JOIN articles_thumbnails at
    ON a.id = at.article_id
WHERE uf.user_login = $1
`
	getAllUnreadUserArticleIds = `
SELECT a.id
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
LEFT OUTER JOIN users_articles_states uas
	ON a.id = uas.article_id AND uf.user_login = uas.user_login
WHERE uas.article_id IS NULL OR NOT uas.read
`
	getAllFavoriteUserArticleIds = `
SELECT a.id
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
LEFT OUTER JOIN users_articles_states uas
	ON a.id = uas.article_id AND uf.user_login = uas.user_login
WHERE uas.favorite
`
	getUserArticleCount = `
SELECT count(a.id)
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
`
	createMissingUserArticleStateByDate = `
INSERT INTO users_articles_states (user_login, article_id)
SELECT uf.user_login, a.id
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
	AND a.id IN (
		SELECT id FROM articles where date IS NULL OR date < $2
	)
EXCEPT SELECT uas.user_login, uas.article_id
FROM users_articles_states uas
WHERE uas.user_login = $1
`
	updateAllUserArticleReadStateByDate = `
UPDATE users_articles_states SET read = $1 WHERE user_login = $2 AND article_id IN (
	SELECT id FROM articles WHERE date IS NULL OR date < $3
)
`
	createNewerMissingUserArticleStateByDate = `
INSERT INTO users_articles_states (user_login, article_id)
SELECT uf.user_login, a.id
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
	AND a.id IN (
		SELECT id FROM articles where date > $2
	)
EXCEPT SELECT uas.user_login, uas.article_id
FROM users_articles_states uas
WHERE uas.user_login = $1
`
	updateAllNewerUserArticleReadStateByDate = `
UPDATE users_articles_states SET read = $1 WHERE user_login = $2 AND article_id IN (
	SELECT id FROM articles WHERE date > $3
)
`
	getUserUnreadCount = `
SELECT count(a.id)
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id
	AND uf.user_login = $1
LEFT OUTER JOIN users_articles_states uas
	ON a.id = uas.article_id AND uf.user_login = uas.user_login
WHERE uas.article_id IS NULL OR NOT uas.read
`
)
