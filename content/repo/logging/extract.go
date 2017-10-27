package logging

import (
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type extractRepo struct {
	repo.Extract

	log log.Log
}

func (r extractRepo) Get(article content.Article) (content.Extract, error) {
	start := time.Now()

	extract, err := r.Extract.Get(article)

	r.log.Infof("repo.Extract.Get took %s", time.Now().Sub(start))

	return extract, err
}

func (r extractRepo) Update(extract content.Extract) error {
	start := time.Now()

	err := r.Extract.Update(extract)

	r.log.Infof("repo.Extract.Update took %s", time.Now().Sub(start))

	return err
}
