package dgraph

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/dgo"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/log"
)

const (
	tagPredicates = `uid tag.value`
)

type Tag content.Tag

func (t Tag) MarshalJSON() ([]byte, error) {
	res := tagInter{
		UID:   NewUID(int64(t.ID)),
		Value: t.Value,
	}

	return json.Marshal(res)
}

func (t *Tag) UnmarshalJSON(b []byte) error {
	res := tagInter{}
	if err := json.Unmarshal(b, &res); err != nil {
		return errors.WithMessage(err, "unmarshaling intermediate tag data")
	}

	t.ID = content.TagID(res.UID.ToInt())
	t.Value = res.Value

	return nil
}

type tagInter struct {
	UID
	Value content.TagValue `json:"tag.value"`
}

type tagRepo struct {
	dg *dgo.Dgraph

	log log.Log
}

func (r tagRepo) Get(id content.TagID, user content.User) (content.Tag, error) {
	if err := user.Validate(); err != nil {
		return content.Tag{}, errors.WithMessage(err, "validating user")
	}

	r.log.Infof("Getting tag %d for %s", id, user)

	query := `query Tag($id: string, $login: string) {
tags as var(func: uid($id)) @cascade {
	uid tag.value
	~tag @filter(eq(login, $login)) {
		uid
	}
}

tags(func: uid(tags)) {
	%s
}
	}`

	resp, err := r.dg.NewReadOnlyTxn().QueryWithVars(
		context.Background(), fmt.Sprintf(query, tagPredicates),
		map[string]string{"$id": intToUid(int64(id)), "$login": string(user.Login)},
	)
	if err != nil {
		return content.Tag{}, errors.Wrapf(err, "getting tag %d for user %s", id, user)
	}

	var root struct {
		Tags []Tag `json:"tags"`
	}

	if err := json.Unmarshal(resp.Json, &root); err != nil {
		return content.Tag{}, errors.Wrap(err, "unmarshaling tag data")
	}

	if len(root.Tags) == 0 {
		return content.Tag{}, content.ErrNoContent
	}

	return content.Tag(root.Tags[0]), nil
}

func (r tagRepo) ForUser(user content.User) ([]content.Tag, error) {
	if err := user.Validate(); err != nil {
		return nil, errors.WithMessage(err, "validating user")
	}

	r.log.Infof("Getting user %s tags", user)

	query := `query Tag($login: string) {
tags as var(func: has(tag.value)) @cascade {
	uid tag.value
	~tag @filter(eq(login, $login)) {
		uid
	}
}

tags(func: uid(tags)) {
	%s
}
	}`

	resp, err := r.dg.NewReadOnlyTxn().QueryWithVars(
		context.Background(), fmt.Sprintf(query, tagPredicates),
		map[string]string{"$login": string(user.Login)},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "getting tags for user %s", user)
	}

	var root struct {
		Tags []Tag `json:"tags"`
	}

	if err := json.Unmarshal(resp.Json, &root); err != nil {
		return nil, errors.Wrap(err, "unmarshaling tag data")
	}

	tags := make([]content.Tag, len(root.Tags))
	for i := range root.Tags {
		tags[i] = content.Tag(root.Tags[i])
	}

	return tags, nil
}

func (r tagRepo) ForFeed(feed content.Feed, user content.User) ([]content.Tag, error) {
	if err := user.Validate(); err != nil {
		return []content.Tag{}, errors.WithMessage(err, "validating user")
	}

	if err := feed.Validate(); err != nil {
		return []content.Tag{}, errors.WithMessage(err, "validating feed")
	}

	r.log.Infof("Getting tags for user %s feed %s", user, feed)

	query := `query Tag($id: string, $login: string) {
tags as var(func: has(tag.value)) @cascade {
	uid tag.value
	~tag @filter(eq(login, $login)) {
		uid
	}
	feed @filter(uid($id)) {
		uid
	}
}

tags(func: uid(tags)) {
	%s
}
	}`

	resp, err := r.dg.NewReadOnlyTxn().QueryWithVars(
		context.Background(), fmt.Sprintf(query, tagPredicates),
		map[string]string{"$id": intToUid(int64(feed.ID)), "$login": string(user.Login)},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "getting tags for user %s and feed %s", user, feed)
	}

	var root struct {
		Tags []Tag `json:"tags"`
	}

	if err := json.Unmarshal(resp.Json, &root); err != nil {
		return nil, errors.Wrap(err, "unmarshaling tag data")
	}

	tags := make([]content.Tag, len(root.Tags))
	for i := range root.Tags {
		tags[i] = content.Tag(root.Tags[i])
	}

	return tags, nil
}

func (r tagRepo) FeedIDs(tag content.Tag, user content.User) ([]content.FeedID, error) {
	if err := tag.Validate(); err != nil {
		return nil, errors.WithMessage(err, "validating tag")
	}

	if err := user.Validate(); err != nil {
		return nil, errors.WithMessage(err, "validating user")
	}

	r.log.Infof("Getting tag %s feed ids", tag)

	query := `query IDs($id: string, $login: string) {
feeds as var(func: has(feed.link)) @cascade {
	uid feed.link
	~feed @filter(eq(login, $login))
}

forTags as var(func: uid(feeds)) @cascade {
	~feed @filter(uid($id))
}

ids(func: uid(forTags)) {
	uid
}
	}`

	resp, err := r.dg.NewReadOnlyTxn().QueryWithVars(
		context.Background(), query,
		map[string]string{"$id": intToUid(int64(tag.ID)), "$login": string(user.Login)},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "getting feed ids for tag %s and user %s", tag, user)
	}

	var root struct {
		UIDs []UID `json:"ids"`
	}

	if err := json.Unmarshal(resp.Json, &root); err != nil {
		return nil, errors.Wrap(err, "unmarshaling feed data")
	}

	ids := make([]content.FeedID, 0, len(root.UIDs))

	for _, uid := range root.UIDs {
		ids = append(ids, content.FeedID(uid.ToInt()))
	}

	return ids, nil
}

func tagUID(ctx context.Context, tx *dgo.Txn, t content.Tag) (UID, error) {
	resp, err := tx.QueryWithVars(ctx, `
query Uid($value: string) {
	uid(func: eq(tag.value, $value)) {
		uid
	}
}`, map[string]string{"$value": string(t.Value)})
	if err != nil {
		return UID{}, errors.Wrapf(err, "querying for existing tag %s", t)
	}

	var data struct {
		UID []UID `json:"uid"`
	}

	if err := json.Unmarshal(resp.Json, &data); err != nil {
		return UID{}, errors.Wrapf(err, "parsing tag query for %s", t)
	}

	if len(data.UID) == 0 {
		return UID{}, errors.Wrapf(content.ErrNoContent, "tag %s not found", t)
	}

	return data.UID[0], nil
}
