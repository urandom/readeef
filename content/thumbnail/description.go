package thumbnail

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type description struct {
	repo repo.Thumbnail
	log  log.Log
}

func FromDescription(repo repo.Thumbnail, log log.Log) Generator {
	return description{repo: repo, log: log}
}

func (t description) Generate(a content.Article) error {
	thumbnail := content.Thumbnail{ArticleID: a.ID, Processed: true}

	t.log.Debugf("Generating thumbnail for article %s from description", a)

	thumbnail.Thumbnail, thumbnail.Link =
		generateThumbnailFromDescription(strings.NewReader(a.Description))

	if err := t.repo.Update(thumbnail); err != nil {
		return errors.WithMessage(err, fmt.Sprintf("saving thumbnail of %s", a))
	}

	return nil
}
