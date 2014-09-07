package readeef

const (
	get_hubbub_subscription = `
SELECT feed_id, lease_duration, verification_time, subscription_failure
	FROM hubbub_subscriptions WHERE link = $1`

	get_hubbub_subscription_by_feed = `
SELECT link, lease_duration, verification_time, subscription_failure
	FROM hubbub_subscriptions WHERE feed_id = $1`

	create_hubbub_subscription = `
INSERT INTO hubbub_subscriptions(link, feed_id, lease_duration, verification_time, subscription_failure)
	SELECT $1, $2, $3, $4, $5 EXCEPT
	SELECT link, feed_id, lease_duration, verification_time, subscription_failure
		FROM hubbub_subscriptions WHERE link = $1
`

	update_hubbub_subscription = `
UPDATE hubbub_subscriptions SET feed_id = $1, lease_duration = $2,
	verification_time = $3, subscription_failure = $4 WHERE link = $5
`

	delete_hubbub_subscription = `DELETE from hubbub_subscriptions where link = $1`

	get_hubbub_subscriptions = `
SELECT link, feed_id, lease_duration, verification_time, subscription_failure
	FROM hubbub_subscriptions`
)

func (db DB) GetHubbubSubscription(link string) (*HubbubSubscription, error) {
	var s *HubbubSubscription

	if err := db.Get(s, db.NamedSQL("get_hubbub_subscription"), link); err != nil {
		return s, err
	}

	s.Link = link

	return s, nil
}

func (db DB) GetHubbubSubscriptionByFeed(feedId int64) (*HubbubSubscription, error) {
	var s *HubbubSubscription

	if err := db.Get(s, db.NamedSQL("get_hubbub_subscription_by_feed"), feedId); err != nil {
		return s, err
	}

	s.FeedId = feedId

	return s, nil
}

func (db DB) UpdateHubbubSubscription(s *HubbubSubscription) error {
	if err := s.Validate(); err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ustmt, err := tx.Preparex(db.NamedSQL("update_hubbub_subscription"))

	if err != nil {
		return err
	}
	defer ustmt.Close()

	res, err := ustmt.Exec(s.FeedId, s.LeaseDuration, s.VerificationTime, s.SubscriptionFailure, s.Link)
	if err != nil {
		return err
	}

	if num, err := res.RowsAffected(); err == nil && num > 0 {
		tx.Commit()
		return nil
	}

	cstmt, err := tx.Preparex(db.NamedSQL("create_hubbub_subscription"))

	if err != nil {
		return err
	}
	defer cstmt.Close()

	_, err = cstmt.Exec(s.Link, s.FeedId, s.LeaseDuration, s.VerificationTime, s.SubscriptionFailure)
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func (db DB) DeleteHubbubSubscription(s *HubbubSubscription) error {
	if err := s.Validate(); err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.NamedSQL("delete_hubbub_subscription"))

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(s.Link)
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func (db DB) GetHubbubSubscriptions() ([]*HubbubSubscription, error) {
	var s []*HubbubSubscription

	if err := db.Select(&s, db.NamedSQL("get_hubbub_subscriptions")); err != nil {
		return s, err
	}

	return s, nil
}

func init() {
	sql_stmt["generic:get_hubbub_subscription"] = get_hubbub_subscription
	sql_stmt["generic:get_hubbub_subscription_by_feed"] = get_hubbub_subscription_by_feed
	sql_stmt["generic:create_hubbub_subscription"] = create_hubbub_subscription
	sql_stmt["generic:update_hubbub_subscription"] = update_hubbub_subscription
	sql_stmt["generic:delete_hubbub_subscription"] = delete_hubbub_subscription
	sql_stmt["generic:get_hubbub_subscriptions"] = get_hubbub_subscriptions
}
