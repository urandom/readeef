package dgraph

import (
	"context"

	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"github.com/pkg/errors"
)

const (
	schema = `
login: string @index(hash) @upsert .
md5api: string @index(hash) .
admin: bool .
active: bool .
feed: uid @reverse .
tag: uid @reverse .

feed.link: string @index(hash) .
ttl: int .
subscription: uid .

tag.value: string @index(hash) @upsert .

sub.link: string @index(hash) @upsert .
leaseDuration: int .
verificationTime: dateTime .
subscriptionFailure: bool .

article: uid .
score: uid .
total.score: int .
score1: int .
score2: int .
score3: int .
score4: int .
score5: int .
score6: int .

guid: string @index(hash) .
date: dateTime .
`
)

func loadSchema(dg *dgo.Dgraph) error {
	op := api.Operation{Schema: schema}

	if err := dg.Alter(context.Background(), &op); err != nil {
		return errors.Wrap(err, "altering schema")
	}

	return nil
}
