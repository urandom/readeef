package ql

const (
	getFeed              = `SELECT link, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds WHERE id() = $1`
	getFeedByLink        = `SELECT id(), title, description, hub_link, site_link, update_error, subscribe_error FROM feeds WHERE link = $1`
	getFeeds             = `SELECT id(), link, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds`
	getUnsubscribedFeeds = `
SELECT id(f), f.link, f.title, f.description, f.hub_link, f.site_link, f.update_error, f.subscribe_error
	FROM feeds f LEFT OUTER JOIN hubbub_subscriptions hs
	ON f.id = hs.feed_id AND hs.subscription_failure = '1'
	ORDER BY f.title
`
)
