package base

func init() {
	sqlStmts.Feed.Get = getFeed
	sqlStmts.Feed.GetByLink = getFeedByLink
	sqlStmts.Feed.GetForUser = getUserFeed
	sqlStmts.Feed.All = getFeeds
	sqlStmts.Feed.AllForUser = getUserFeeds
	sqlStmts.Feed.AllForTag = getUserTagFeeds
	sqlStmts.Feed.Unsubscribed = getUnsubscribedFeeds

	sqlStmts.Feed.IDs = feedIDs
	sqlStmts.Feed.Create = createFeed
	sqlStmts.Feed.Update = updateFeed
	sqlStmts.Feed.Delete = deleteFeed
	sqlStmts.Feed.GetLatestArticles = getLatestFeedArticles
	sqlStmts.Feed.GetUsers = getFeedUsers
	sqlStmts.Feed.Attach = createUserFeed
	sqlStmts.Feed.Detach = deleteUserFeed
	sqlStmts.Feed.CreateUserTag = createUserFeedTag
	sqlStmts.Feed.DeleteUserTags = deleteUserFeedTags
}

const (
	feedIDs    = `SELECT id FROM feeds`
	createFeed = `
INSERT INTO feeds(link, title, description, hub_link, site_link, update_error, subscribe_error)
	SELECT $1, $2, $3, $4, $5, $6, $7 EXCEPT SELECT link, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds WHERE link = $1`
	updateFeed = `UPDATE feeds SET link = $1, title = $2, description = $3, hub_link = $4, site_link = $5, update_error = $6, subscribe_error = $7 WHERE id = $8`
	deleteFeed = `DELETE FROM feeds WHERE id = $1`

	getLatestFeedArticles = `
SELECT a.feed_id, a.id, a.title, a.description, a.link, a.date, a.guid
FROM articles a
WHERE a.feed_id = $1 AND a.date > NOW() - INTERVAL '5 days'
`
	getFeedUsers = `
SELECT u.login, u.first_name, u.last_name, u.email, u.admin, u.active,
	u.profile_data, u.hash_type, u.salt, u.hash, u.md5_api
FROM users u, users_feeds uf
WHERE u.login = uf.user_login AND uf.feed_id = $1
`
	createUserFeed = `
INSERT INTO users_feeds(user_login, feed_id)
	SELECT $1, $2 EXCEPT SELECT user_login, feed_id FROM users_feeds WHERE user_login = $1 AND feed_id = $2`
	deleteUserFeed = `DELETE FROM users_feeds WHERE user_login = $1 AND feed_id = $2`

	createUserFeedTag = `
INSERT INTO users_feeds_tags(user_login, feed_id, tag_id)
	SELECT $1, $2, $3 EXCEPT SELECT user_login, feed_id, tag_id
		FROM users_feeds_tags
		WHERE user_login = $1 AND feed_id = $2 AND tag_id = $3
`
	deleteUserFeedTags = `
DELETE FROM users_feeds_tags WHERE user_login = $1 AND feed_id = $2
`

	getFeed       = `SELECT link, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds WHERE id = $1`
	getFeedByLink = `SELECT id, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds WHERE link = $1`
	getUserFeed   = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds uf
WHERE f.id = uf.feed_id
	AND f.id = $1 AND uf.user_login = $2
`
	getFeeds     = `SELECT id, link, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds`
	getUserFeeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds uf
WHERE f.id = uf.feed_id
	AND uf.user_login = $1
ORDER BY LOWER(f.title)
`
	getUserTagFeeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds_tags uft, tags t
WHERE f.id = uft.feed_id
	AND t.id = uft.tag_id
	AND uft.user_login = $1 AND t.value = $2
ORDER BY LOWER(f.title)
`
	getUnsubscribedFeeds = `
SELECT f.id, f.link, f.title, f.description, f.hub_link, f.site_link, f.update_error, f.subscribe_error
	FROM feeds f LEFT OUTER JOIN hubbub_subscriptions hs
	ON f.id = hs.feed_id AND hs.subscription_failure = '1'
	ORDER BY f.title
`
)
