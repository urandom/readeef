package logging

import (
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type thumbnailRepo struct {
	repo.Thumbnail

	log log.Log
}

func (r thumbnailRepo) Get(article content.Article) (content.Thumbnail, error) {
	start := time.Now()

	thumbnail, err := r.Thumbnail.Get(article)

	r.log.Infof("repo.Thumbnail.Get took %s", time.Now().Sub(start))

	return thumbnail, err
}

func (r thumbnailRepo) Update(thumbnail content.Thumbnail) error {
	start := time.Now()

	err := r.Thumbnail.Update(thumbnail)

	r.log.Infof("repo.Thumbnail.Update took %s", time.Now().Sub(start))

	return err
}
