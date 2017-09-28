package base

func init() {
	sqlStmts.Subscription.Create = createHubbubSubscription
	sqlStmts.Subscription.Update = updateHubbubSubscription
	sqlStmts.Subscription.Delete = deleteHubbubSubscription
}

const (
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
	deleteHubbubSubscription = `DELETE from hubbub_subscriptions where feed_id = $1`
)
