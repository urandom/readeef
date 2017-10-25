package logging

import (
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type articleRepo struct {
	repo.Article

	log log.Log
}

func (r articleRepo) ForUser(user content.User, opts ...content.QueryOpt) ([]content.Article, error) {
	start := time.Now()

	articles, err := r.Article.ForUser(user, opts...)

	r.log.Infof("repo.Article.ForUser took %s", time.Now().Sub(start))

	return articles, err
}

func (r articleRepo) All(opts ...content.QueryOpt) ([]content.Article, error) {
	start := time.Now()

	articles, err := r.Article.All(opts...)

	r.log.Infof("repo.Article.All took %s", time.Now().Sub(start))

	return articles, err
}

func (r articleRepo) Count(user content.User, opts ...content.QueryOpt) (int64, error) {
	start := time.Now()

	count, err := r.Article.Count(user, opts...)

	r.log.Infof("repo.Article.Count took %s", time.Now().Sub(start))

	return count, err
}

func (r articleRepo) IDs(user content.User, opts ...content.QueryOpt) ([]content.ArticleID, error) {
	start := time.Now()

	ids, err := r.Article.IDs(user, opts...)

	r.log.Infof("repo.Article.IDs took %s", time.Now().Sub(start))

	return ids, err
}

func (r articleRepo) Read(state bool, user content.User, opts ...content.QueryOpt) error {
	start := time.Now()

	err := r.Article.Read(state, user, opts...)

	r.log.Infof("repo.Article.Read took %s", time.Now().Sub(start))

	return err
}

func (r articleRepo) Favor(state bool, user content.User, opts ...content.QueryOpt) error {
	start := time.Now()

	err := r.Article.Favor(state, user, opts...)

	r.log.Infof("repo.Article.Favor took %s", time.Now().Sub(start))

	return err
}

func (r articleRepo) RemoveStaleUnreadRecords() error {
	start := time.Now()

	err := r.Article.RemoveStaleUnreadRecords()

	r.log.Infof("repo.Article.RemoveStaleUnreadRecords took %s", time.Now().Sub(start))

	return err
}
