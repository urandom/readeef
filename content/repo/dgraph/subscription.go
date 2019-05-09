package dgraph

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/log"
)

const (
	subscriptionPredicates = `sub.link leaseDuration verificationTime subscriptionFailure`
)

type Subscription struct {
	content.Subscription
	UID
}

type subscriptionInter struct {
	UID
	FeedID              string    `json:"feed,omitempty"`
	Link                string    `json:"sub.link"`
	LeaseDuration       int64     `json:"leaseDuration"`
	VerificationTime    time.Time `json:"verificationTime"`
	SubscriptionFailure bool      `json:"subscriptionFailure"`
}

func (s Subscription) MarshalJSON() ([]byte, error) {
	res := subscriptionInter{
		UID:                 s.UID,
		Link:                s.Link,
		LeaseDuration:       s.LeaseDuration,
		VerificationTime:    s.VerificationTime,
		SubscriptionFailure: s.SubscriptionFailure,
	}

	return json.Marshal(res)
}

func (s *Subscription) UnmarshalJSON(b []byte) error {
	res := subscriptionInter{}
	if err := json.Unmarshal(b, &res); err != nil {
		return errors.WithMessage(err, "unmarshaling intermediate subscription data")
	}

	if res.FeedID != "" {
		s.FeedID = content.FeedID(UID{Value: res.FeedID}.ToInt())
	}
	s.Link = res.Link
	s.LeaseDuration = res.LeaseDuration
	s.VerificationTime = res.VerificationTime
	s.SubscriptionFailure = res.SubscriptionFailure

	return nil
}

type subscriptionRepo struct {
	dg *dgo.Dgraph

	log log.Log
}

func (r subscriptionRepo) Get(feed content.Feed) (content.Subscription, error) {
	if err := feed.Validate(); err != nil {
		return content.Subscription{}, errors.WithMessage(err, "validating feed")
	}

	r.log.Infoln("Getting feed subscription")

	query := `query S($id: string) {
var(func: uid($id)) @cascade {
	sub as subscription {
		sub.link
	}
}

subscriptions(func: uid(sub)) {
	%s
}
	}`

	resp, err := r.dg.NewReadOnlyTxn().QueryWithVars(
		context.Background(), fmt.Sprintf(query, tagPredicates),
		map[string]string{"$id": intToUid(int64(feed.ID))},
	)
	if err != nil {
		return content.Subscription{}, errors.Wrapf(err, "getting subscription for feed %s", feed)
	}

	var root struct {
		Subscriptions []Subscription `json:"subscriptions"`
	}

	if err := json.Unmarshal(resp.Json, &root); err != nil {
		return content.Subscription{}, errors.Wrap(err, "unmarshaling subscription data")
	}

	if len(root.Subscriptions) == 0 {
		return content.Subscription{}, content.ErrNoContent
	}

	s := root.Subscriptions[0].Subscription
	s.FeedID = feed.ID

	return s, nil
}

func (r subscriptionRepo) All() ([]content.Subscription, error) {
	r.log.Infoln("Getting all subscriptions")

	query := `{
feeds as var(func: has(subscription)) @cascade {
	subscription {
		sub.link
	}
}

subscriptions(func: uid(feeds)) @normalize {
	feed: uid
	subscription {
		%s
	}
}
	}`

	resp, err := r.dg.NewReadOnlyTxn().Query(
		context.Background(), fmt.Sprintf(query, aliasPredicates(subscriptionPredicates)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "getting subscriptions")
	}

	var root struct {
		Subscriptions []Subscription `json:"subscriptions"`
	}

	if err := json.Unmarshal(resp.Json, &root); err != nil {
		return nil, errors.Wrap(err, "unmarshaling subscription data")
	}

	subs := make([]content.Subscription, len(root.Subscriptions))
	for i := range root.Subscriptions {
		subs[i] = root.Subscriptions[i].Subscription
	}

	return subs, nil
}

func (r subscriptionRepo) Update(subscription content.Subscription) error {
	if err := subscription.Validate(); err != nil {
		return errors.WithMessage(err, "validating subscription")
	}

	r.log.Infof("Updating subscription %s", subscription)

	ctx := context.Background()
	tx := r.dg.NewTxn()
	defer tx.Discard(ctx)

	uid, err := subscriptionUID(ctx, tx, subscription)
	if err != nil && !content.IsNoContent(err) {
		return err
	}

	var b []byte
	if uid.Valid() {
		r.log.Infof("Updating subscription %s with uid %d", subscription, uid.ToInt())
		b, err = json.Marshal(Subscription{subscription, uid})
	} else {
		r.log.Infof("Creating subscription %s", subscription)
		b, err = json.Marshal(Subscription{Subscription: subscription})
	}
	if err != nil {
		return errors.Wrapf(err, "marshaling subscription %s", subscription)
	}

	resp, err := tx.Mutate(ctx, &api.Mutation{
		SetJson: b,
	})

	if err != nil {
		return errors.Wrapf(err, "updating subscription %s", subscription)
	}

	if len(resp.Uids) > 0 {
		uid = UID{resp.Uids["blank-0"]}
	}

	// FIXME: check if the feed ID has changed, delete the old link
	if _, err := tx.Mutate(ctx, &api.Mutation{
		Set:       []*api.NQuad{{Subject: NewUID(int64(subscription.FeedID)).Value, Predicate: "subscription", ObjectId: uid.Value}},
		CommitNow: true,
	}); err != nil {
		return errors.Wrapf(err, "Linking feed %d to subscription %s", subscription.FeedID, subscription)
	}

	return nil
}

func subscriptionUID(ctx context.Context, tx *dgo.Txn, s content.Subscription) (UID, error) {
	resp, err := tx.QueryWithVars(ctx, `
query Uid($link: string) {
	uid(func: eq(sub.link, $link)) {
		uid
	}
}`, map[string]string{"$link": string(s.Link)})
	if err != nil {
		return UID{}, errors.Wrapf(err, "querying for existing subscription %s", s)
	}

	var data struct {
		UID []UID `json:"uid"`
	}

	if err := json.Unmarshal(resp.Json, &data); err != nil {
		return UID{}, errors.Wrapf(err, "parsing subscription query for %s", s)
	}

	if len(data.UID) == 0 {
		return UID{}, errors.Wrapf(content.ErrNoContent, "subscription %s not found", s)
	}

	return data.UID[0], nil
}
