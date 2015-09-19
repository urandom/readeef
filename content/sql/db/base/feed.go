package base

func init() {
	sqlStmts.Feed.Create = createFeed
	sqlStmts.Feed.Update = updateFeed
	sqlStmts.Feed.Delete = deleteFeed
	sqlStmts.Feed.GetAllArticles = getAllFeedArticles
	sqlStmts.Feed.GetLatestArticles = getLatestFeedArticles
	sqlStmts.Feed.GetHubbubSubscription = getFeedHubbubSubscription
	sqlStmts.Feed.GetUsers = getFeedUsers
	sqlStmts.Feed.Detach = deleteUserFeed
	sqlStmts.Feed.GetUserTags = getUserFeedTags
	sqlStmts.Feed.CreateUserTag = createUserFeedTag
	sqlStmts.Feed.DeleteUserTags = deleteUserFeedTags
	sqlStmts.Feed.ReadStateInsertTemplate = readStateInsertFeedTemplate
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
	getFeedHubbubSubscription = `
SELECT link, lease_duration, verification_time, subscription_failure
	FROM hubbub_subscriptions WHERE feed_id = $1`
	getFeedUsers = `
SELECT u.login, u.first_name, u.last_name, u.email, u.admin, u.active,
	u.profile_data, u.hash_type, u.salt, u.hash, u.md5_api
FROM users u, users_feeds uf
WHERE u.login = uf.user_login AND uf.feed_id = $1
`
	deleteUserFeed = `DELETE FROM users_feeds WHERE user_login = $1 AND feed_id = $2`

	getUserFeedTags = `
SELECT t.id, t.value
FROM users_feeds_tags uft INNER JOIN tags t
	ON uft.tag_id = t.id
WHERE uft.user_login = $1 AND uft.feed_id = $2`
	createUserFeedTag = `
INSERT INTO users_feeds_tags(user_login, feed_id, tag_id)
	SELECT $1, $2, $3 EXCEPT SELECT user_login, feed_id, tag_id
		FROM users_feeds_tags
		WHERE user_login = $1 AND feed_id = $2 AND tag_id = $3
`
	deleteUserFeedTags = `
DELETE FROM users_feeds_tags WHERE user_login = $1 AND feed_id = $2
`
	readStateInsertFeedTemplate = `
INSERT INTO users_articles_unread (user_login, article_id)
SELECT uf.user_login, a.id
FROM users_feeds uf
INNER JOIN articles a
	ON uf.feed_id = a.feed_id
	AND uf.feed_id = $1
	AND a.id IN (
		{{ .NewArticleIds }}
	)
EXCEPT SELECT au.user_login, au.article_id
FROM users_articles_unread au
`
)
