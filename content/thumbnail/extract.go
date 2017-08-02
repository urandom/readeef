package thumbnail

import (
	_ "image/png"
	"strings"

	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

type extract struct {
	extractor content.Extractor
	log       readeef.Logger
}

func FromExtract(e content.Extractor, l readeef.Logger) (content.Thumbnailer, error) {
	if e == nil {
		return nil, errors.New("A valid extractor is required")
	}

	return extract{extractor: e, log: l}, nil
}

func (t extract) Generate(a content.Article) error {
	ad := a.Data()

	thumbnail := a.Repo().ArticleThumbnail()
	td := data.ArticleThumbnail{
		ArticleId: ad.Id,
		Processed: true,
	}

	t.log.Debugf("Generating thumbnail for article %s\n", a)

	td.Thumbnail, td.Link =
		generateThumbnailFromDescription(strings.NewReader(ad.Description))

	if td.Link == "" {
		t.log.Debugf("%s description doesn't contain suitable link, getting extract\n", a)

		extract := a.Extract()
		if a.HasErr() {
			return a.Err()
		}

		extractData := extract.Data()

		if extract.HasErr() {
			switch err := extract.Err(); err {
			case content.ErrNoContent:
				t.log.Debugf("Generating article extract for %s\n", a)
				extractData, err = t.extractor.Extract(a.Data().Link)
				if err != nil {
					return err
				}

				extractData.ArticleId = a.Data().Id
				extract.Data(extractData)
				extract.Update()
				if extract.HasErr() {
					return extract.Err()
				}
			default:
				return err
			}
		}

		if extractData.TopImage == "" {
			t.log.Debugf("Extract for %s doesn't contain a top image\n", a)
		} else {
			t.log.Debugf("Generating thumbnail from top image %s of %s\n", extractData.TopImage, a)
			td.Thumbnail = generateThumbnailFromImageLink(extractData.TopImage)
			td.Link = extractData.TopImage
		}
	}

	thumbnail.Data(td)
	if thumbnail.Update(); thumbnail.HasErr() {
		return errors.Wrapf(thumbnail.Err(), "saving thumbnail of %s to database", a)
	}

	return nil
}
