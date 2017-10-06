package base

func init() {
	sqlStmts.Subscription.GetForFeed = getFeedHubbubSubscription
	sqlStmts.Subscription.All = getHubbubSubscriptions
	sqlStmts.Subscription.Create = createHubbubSubscription
	sqlStmts.Subscription.Update = updateHubbubSubscription
}

const (
	getFeedHubbubSubscription = `
SELECT link, lease_duration, verification_time, subscription_failure
	FROM hubbub_subscriptions WHERE feed_id = $1`
	getHubbubSubscriptions = `
SELECT link, feed_id, lease_duration, verification_time, subscription_failure
	FROM hubbub_subscriptions`

	createHubbubSubscription = `
INSERT INTO hubbub_subscriptions(feed_id, link, lease_duration, verification_time, subscription_failure)
	SELECT $1, $2, $3, $4, $5 EXCEPT
	SELECT feed_id, link, lease_duration, verification_time, subscription_failure
		FROM hubbub_subscriptions WHERE feed_id = $1
`
	updateHubbubSubscription = `
UPDATE hubbub_subscriptions SET link = $1, lease_duration = $2,
	verification_time = $3, subscription_failure = $4 WHERE feed_id = $5
`
)
