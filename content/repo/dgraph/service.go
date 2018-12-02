package dgraph

import (
	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
	"google.golang.org/grpc"
)

type Service struct {
	user         repo.User
	tag          repo.Tag
	feed         repo.Feed
	subscription repo.Subscription
	article      repo.Article
	extract      repo.Extract
	scores       repo.Scores
	thumbnail    repo.Thumbnail
}

func NewService(source string, log log.Log) (Service, error) {
	conn, err := grpc.Dial(source, grpc.WithInsecure())
	if err != nil {
		return Service{}, errors.Wrap(err, "creating dgraph service")
	}

	dc := api.NewDgraphClient(conn)
	dg := dgo.NewDgraphClient(dc)

	if err := loadSchema(dg); err != nil {
		return Service{}, err
	}

	return Service{
		user:         userRepo{dg, log},
		feed:         feedRepo{dg, log},
		tag:          tagRepo{dg, log},
		subscription: subscriptionRepo{dg, log},
	}, nil
}

func (s Service) UserRepo() repo.User {
	return s.user
}

func (s Service) TagRepo() repo.Tag {
	return s.tag
}

func (s Service) FeedRepo() repo.Feed {
	return s.feed
}

func (s Service) SubscriptionRepo() repo.Subscription {
	return s.subscription
}

func (s Service) ArticleRepo() repo.Article {
	return s.article
}

func (s Service) ExtractRepo() repo.Extract {
	return s.extract
}

func (s Service) ScoresRepo() repo.Scores {
	return s.scores
}

func (s Service) ThumbnailRepo() repo.Thumbnail {
	return s.thumbnail
}
