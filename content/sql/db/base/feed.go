package base

func init() {
	sql["create_feed"] = createFeed
	sql["update_feed"] = updateFeed
	sql["delete_feed"] = deleteFeed
	sql["get_all_feed_articles"] = getAllFeedArticles
	sql["get_latest_feed_articles"] = getLatestFeedArticles
	sql["get_hubbub_subscription"] = getHubbubSubscription
	sql["get_feed_users"] = getFeedUsers
	sql["delete_user_feed"] = deleteUserFeed
	sql["create_missing_user_article_state_by_feed_date"] = createMissingUserArticleStateByFeedDate
	sql["update_all_user_article_state_by_feed_date"] = updateAllUserArticleReadStateByFeedDate
	sql["create_user_feed_tag"] = createUserFeedTag
	sql["delete_user_feed_tags"] = deleteUserFeedTags
	sql["get_user_feed_tags"] = getUserFeedTags
	sql["get_user_feed_unread_count"] = getUserFeedUnreadCount
}

const (
	createFeed = `
INSERT INTO feeds(link, title, description, hub_link, site_link, update_error, subscribe_error)
	SELECT $1, $2, $3, $4, $5, $6, $7 EXCEPT SELECT link, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds WHERE link = $1`
	updateFeed = `UPDATE feeds SET link = $1, title = $2, description = $3, hub_link = $4, site_link = $5, update_error = $6, subscribe_error = $7 WHERE id = $8`
	deleteFeed = `DELETE FROM feeds WHERE id = $1`

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
	deleteUserFeed = `DELETE FROM users_feeds WHERE user_login = $1 AND feed_id = $2`

	createMissingUserArticleStateByFeedDate = `
INSERT INTO users_articles_states (user_login, article_id)
SELECT uf.user_login, a.id
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1 AND uf.feed_id = $2
	AND a.id IN (
		SELECT id FROM articles where date IS NULL OR date < $3
	)
EXCEPT SELECT uas.user_login, uas.article_id
FROM articles a INNER JOIN users_articles_states uas
	ON a.id = uas.article_id
WHERE uas.user_login = $1 AND a.feed_id = $2
`
	updateAllUserArticleReadStateByFeedDate = `
UPDATE users_articles_states SET read = $1 WHERE user_login = $2 AND article_id IN (
	SELECT id FROM articles WHERE feed_id = $3 AND (date IS NULL OR date < $4)
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
	getUserFeedUnreadCount = `
SELECT count(a.id)
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id
	AND uf.user_login = $1
	AND uf.feed_id = $2
LEFT OUTER JOIN users_articles_states uas
	ON a.id = uas.article_id AND uf.user_login = uas.user_login
WHERE uas.article_id IS NULL OR NOT uas.read
`
)
