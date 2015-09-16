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
	sql["get_all_unread_user_article_ids"] = getAllUnreadUserArticleIds
	sql["get_all_favorite_user_article_ids"] = getAllFavoriteUserArticleIds

	sql["get_articles_template"] = getArticlesTemplate
	sql["get_articles_score_join"] = getArticlesScoreJoin

	sql["read_state_insert_template"] = readStateInsertTemplate
	sql["read_state_insert_favorite_join"] = readStateInsertFavoriteJoin

	sql["read_state_delete_template"] = readStateDeleteTemplate
	sql["read_state_delete_favorite_join"] = readStateInsertFavoriteJoin

	sql["article_count_template"] = articleCountTemplate
	sql["article_count_unread_join"] = articleCountUnreadJoin
	sql["article_count_favorite_join"] = articleCountFavoriteJoin

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

	getAllUnreadUserArticleIds = `
SELECT a.id
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
LEFT OUTER JOIN users_articles_unread au
	ON a.id = au.article_id AND uf.user_login = au.user_login
WHERE au.article_id IS NOT NULL
`
	getAllFavoriteUserArticleIds = `
SELECT a.id
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
LEFT OUTER JOIN users_articles_favorite af
	ON a.id = af.article_id AND uf.user_login = af.user_login
WHERE af.article_id IS NOT NULL
`

	getArticlesTemplate = `
SELECT a.feed_id, a.id, a.title, a.description, a.link, a.date, a.guid,
	CASE WHEN au.article_id IS NULL THEN 1 ELSE 0 END AS read,
	CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite,
	COALESCE(at.thumbnail, '') as thumbnail,
	COALESCE(at.link, '') as thumbnail_link
	{{ .Columns }}
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id
{{ .Join }}
LEFT OUTER JOIN users_articles_unread au 
    ON a.id = au.article_id AND uf.user_login = au.user_login
LEFT OUTER JOIN users_articles_favorite af 
    ON a.id = af.article_id AND uf.user_login = af.user_login
LEFT OUTER JOIN articles_thumbnails at
    ON a.id = at.article_id
WHERE uf.user_login = $1
{{ .Where }}
{{ .Order }}
{{ .Limit }}
`
	getArticlesScoreJoin = `
	INNER JOIN articles_scores asco ON a.id = asco.article_id
`

	readStateInsertTemplate = `
INSERT INTO users_articles_unread (user_login, article_id)
SELECT uf.user_login, a.id
FROM users_feeds uf
INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
	{{ .JoinPredicate }}
{{ .Join }}
{{ .Where }}
EXCEPT SELECT au.user_login, au.article_id
FROM users_articles_unread au
WHERE au.user_login = $1
`
	readStateInsertFavoriteJoin = `
LEFT OUTER JOIN users_articles_favorite af
	ON a.id = af.article_id AND af.user_login = uf.user_login
`

	readStateDeleteTemplate = `
DELETE FROM users_articles_unread WHERE user_login = $1 AND article_id IN (
	SELECT a.id FROM articles a {{ .Join }}
	{{ .Where }}
)
`

	articleCountTemplate = `
SELECT count(a.id)
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id
	AND uf.user_login = $1
{{ .Join }}
{{ .Where }}
`
	articleCountUnreadJoin = `
LEFT OUTER JOIN users_articles_unread au
	ON a.id = au.article_id
	AND uf.user_login = au.user_login
`
	articleCountFavoriteJoin = `
LEFT OUTER JOIN users_articles_favorite af
	ON a.id = af.article_id
	AND uf.user_login = af.user_login
`
)
