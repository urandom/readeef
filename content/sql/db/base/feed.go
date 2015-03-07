package base

func init() {
	sql["create_feed"] = createFeed
	sql["update_feed"] = updateFeed
	sql["delete_feed"] = deleteFeed
	sql["create_feed_article"] = createFeedArticle
	sql["update_feed_article"] = updateFeedArticle
	sql["update_feed_article_with_guid"] = updateFeedArticleWithGuid
	sql["get_all_feed_articles"] = getAllFeedArticles
	sql["get_latest_feed_articles"] = getLatestFeedArticles
	sql["get_hubbub_subscription"] = getHubbubSubscription
	sql["get_feed_users"] = getFeedUsers
	sql["delete_user_feed"] = deleteUserFeed
	sql["create_all_users_articles_read_by_feed_date"] = createAllUsersArticlesReadByFeedDate
	sql["delete_all_users_articles_read_by_feed_date"] = deleteAllUsersArticlesReadByFeedDate
	sql["create_user_feed_tag"] = createUserFeedTag
	sql["delete_user_feed_tags"] = deleteUserFeedTags
	sql["get_user_feed_tags"] = getUserFeedTags
}

const (
	createFeed = `
INSERT INTO feeds(link, title, description, hub_link, site_link, update_error, subscribe_error)
	SELECT $1, $2, $3, $4, $5, $6, $7 EXCEPT SELECT link, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds WHERE link = $1`
	updateFeed        = `UPDATE feeds SET link = $1, title = $2, description = $3, hub_link = $4, site_link = $5, update_error = $6, subscribe_error = $7 WHERE id = $8`
	deleteFeed        = `DELETE FROM feeds WHERE id = $1`
	createFeedArticle = `
INSERT INTO articles(feed_id, link, guid, title, description, date)
	SELECT $1, $2, $3, $4, $5, $6 EXCEPT
		SELECT feed_id, link, guid, title, description, date
		FROM articles WHERE feed_id = $1 AND link = $2
`

	updateFeedArticle = `
UPDATE articles SET title = $1, description = $2, date = $3 WHERE feed_id = $4 AND link = $5
`

	updateFeedArticleWithGuid = `
UPDATE articles SET title = $1, description = $2, date = $3 WHERE feed_id = $4 AND guid = $5
`
	getAllFeedArticles = `
SELECT a.feed_id, a.id, a.title, a.description, a.link, a.guid, a.date
FROM articles a
WHERE a.feed_id = $1
`
	getLatestFeedArticles = `
SELECT a.feed_id, a.id, a.title, a.description, a.link, a.date, a.guid
FROM articles a
WHERE a.feed_id = $1 AND a.date > NOW() - INTERVAL '5 days'
`
	getHubbubSubscription = `
SELECT link, lease_duration, verification_time, subscription_failure
	FROM hubbub_subscriptions WHERE feed_id = $1`
	getFeedUsers = `
SELECT u.login, u.first_name, u.last_name, u.email, u.admin, u.active,
	u.profile_data, u.hash_type, u.salt, u.hash, u.md5_api
FROM users u, users_feeds uf
WHERE u.login = uf.user_login AND uf.feed_id = $1
`
	deleteUserFeed                       = `DELETE FROM users_feeds WHERE user_login = $1 AND feed_id = $2`
	createAllUsersArticlesReadByFeedDate = `
INSERT INTO users_articles_read
	SELECT uf.user_login, a.id
	FROM users_feeds uf INNER JOIN articles a
		ON uf.feed_id = a.feed_id AND uf.user_login = $1 AND uf.feed_id = $2
		AND a.id IN (SELECT id FROM articles WHERE date IS NULL OR date < $3)
`

	deleteAllUsersArticlesReadByFeedDate = `
DELETE FROM users_articles_read WHERE user_login = $1 AND article_id IN (
	SELECT id FROM articles WHERE feed_id = $2 AND (date IS NULL OR date < $3)
)
`
	getUserFeedTags   = `SELECT tag FROM users_feeds_tags WHERE user_login = $1 AND feed_id = $2`
	createUserFeedTag = `
INSERT INTO users_feeds_tags(user_login, feed_id, tag)
	SELECT $1, $2, $3 EXCEPT SELECT user_login, feed_id, tag
		FROM users_feeds_tags
		WHERE user_login = $1 AND feed_id = $2 AND tag = $3
`
	deleteUserFeedTags = `
DELETE FROM users_feeds_tags WHERE user_login = $1 AND feed_id = $2
`
)
