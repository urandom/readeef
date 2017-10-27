package logging

import (
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type scoresRepo struct {
	repo.Scores

	log log.Log
}

func (r scoresRepo) Get(article content.Article) (content.Scores, error) {
	start := time.Now()

	scores, err := r.Scores.Get(article)

	r.log.Infof("repo.Scores.Get took %s", time.Now().Sub(start))

	return scores, err
}

func (r scoresRepo) Update(scores content.Scores) error {
	start := time.Now()

	err := r.Scores.Update(scores)

	r.log.Infof("repo.Scores.Update took %s", time.Now().Sub(start))

	return err
}
