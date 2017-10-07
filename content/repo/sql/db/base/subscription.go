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
FROM hubbub_subscriptions WHERE feed_id = :feed_id`
	getHubbubSubscriptions = `
SELECT link, feed_id, lease_duration, verification_time, subscription_failure
	FROM hubbub_subscriptions`

	createHubbubSubscription = `
INSERT INTO hubbub_subscriptions(feed_id, link, lease_duration, verification_time, subscription_failure)
	SELECT :feed_id, :link, :lease_duration, :verification_time, :subscription_failure EXCEPT
	SELECT feed_id, link, lease_duration, verification_time, subscription_failure
		FROM hubbub_subscriptions WHERE feed_id = :feed_id
`
	updateHubbubSubscription = `
UPDATE hubbub_subscriptions SET link = :link, lease_duration = :lease_duration,
	verification_time = :verification_time, subscription_failure = :subscription_failure WHERE feed_id = :feed_id 
`
)
