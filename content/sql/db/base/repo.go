package base

func init() {
	sql["get_user"] = getUser
	sql["get_user_by_md5_api"] = getUserByMD5Api
	sql["get_users"] = getUsers
	sql["get_feed"] = getFeed
	sql["get_feed_by_link"] = getFeedByLink
	sql["get_feeds"] = getFeeds
	sql["get_unsubscribed_feeds"] = getUnsubscribedFeeds
	sql["get_hubbub_subscriptions"] = getHubbubSubscriptions
	sql["fail_hubbub_subscriptions"] = failHubbubSubscription
	sql["delete_stale_unread_records"] = deleteStaleUnreadRecords
}

const (
	getUser         = `SELECT first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api FROM users WHERE login = $1`
	getUserByMD5Api = `SELECT login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash FROM users WHERE md5_api = $1`
	getUsers        = `SELECT login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api FROM users`

	getFeed              = `SELECT link, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds WHERE id = $1`
	getFeedByLink        = `SELECT id, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds WHERE link = $1`
	getFeeds             = `SELECT id, link, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds`
	getUnsubscribedFeeds = `
SELECT f.id, f.link, f.title, f.description, f.hub_link, f.site_link, f.update_error, f.subscribe_error
	FROM feeds f LEFT OUTER JOIN hubbub_subscriptions hs
	ON f.id = hs.feed_id AND hs.subscription_failure = '1'
	ORDER BY f.title
`
	getHubbubSubscriptions = `
SELECT link, feed_id, lease_duration, verification_time, subscription_failure
	FROM hubbub_subscriptions`
	failHubbubSubscription   = `UPDATE hubbub_subscriptions SET subscription_failure = '1'`
	deleteStaleUnreadRecords = `DELETE FROM users_articles_unread WHERE insert_date < $1`
)
