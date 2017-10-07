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
SELECT :link, :title, :description, :hub_link, :site_link, :update_error, :subscribe_error EXCEPT SELECT link, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds WHERE link = :link`
	updateFeed = `UPDATE feeds SET link = :link, title = :title, description = :description, hub_link = :hub_link, site_link = :site_link, update_error = :update_error, subscribe_error = :subscribe_error WHERE id = :id`
	deleteFeed = `DELETE FROM feeds WHERE id = :id`

	getFeedUsers = `
SELECT u.login, u.first_name, u.last_name, u.email, u.admin, u.active,
	u.profile_data, u.hash_type, u.salt, u.hash, u.md5_api
FROM users u, users_feeds uf
WHERE u.login = uf.user_login AND uf.feed_id = :id
`
	createUserFeed = `
INSERT INTO users_feeds(user_login, feed_id)
	SELECT :user_login, :id EXCEPT SELECT user_login, feed_id FROM users_feeds
		WHERE user_login = :user_login AND feed_id = :id`
	deleteUserFeed = `DELETE FROM users_feeds WHERE user_login = :user_login AND feed_id = :id`

	createUserFeedTag = `
INSERT INTO users_feeds_tags(user_login, feed_id, tag_id)
	SELECT :user_login, :feed_id, :tag_id EXCEPT SELECT user_login, feed_id, tag_id
		FROM users_feeds_tags
		WHERE user_login = :user_login AND feed_id = :feed_id AND tag_id = :tag_id
`
	deleteUserFeedTags = `
DELETE FROM users_feeds_tags WHERE user_login = :user_login AND feed_id = :feed_id
`

	getFeed       = `SELECT link, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds WHERE id = :id`
	getFeedByLink = `SELECT id, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds WHERE link = :link`
	getUserFeed   = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds uf
WHERE f.id = uf.feed_id
	AND f.id = :id AND uf.user_login = :user_login
`
	getFeeds     = `SELECT id, link, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds`
	getUserFeeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds uf
WHERE f.id = uf.feed_id
	AND uf.user_login = :user_login
ORDER BY LOWER(f.title)
`
	getUserTagFeeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds_tags uft, tags t
WHERE f.id = uft.feed_id
	AND t.id = uft.tag_id
	AND uft.user_login = :user_login AND t.value = :tag_value
ORDER BY LOWER(f.title)
`
	getUnsubscribedFeeds = `
SELECT f.id, f.link, f.title, f.description, f.hub_link, f.site_link, f.update_error, f.subscribe_error
	FROM feeds f LEFT OUTER JOIN hubbub_subscriptions hs
	ON f.id = hs.feed_id AND hs.subscription_failure = '1'
	ORDER BY f.title
`
)
