package thumbnailer

import (
	"fmt"
	_ "image/png"
	"strings"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/webfw"
)

type Extract struct {
	extractor content.Extractor
	logger    webfw.Logger
}

func NewExtract(e content.Extractor, l webfw.Logger) (content.Thumbnailer, error) {
	if e == nil {
		return nil, fmt.Errorf("A valid extractor is required")
	}
	return Extract{extractor: e, logger: l}, nil
}

func (t Extract) Generate(a content.Article) error {
	ad := a.Data()

	thumbnail := a.Repo().ArticleThumbnail()
	td := data.ArticleThumbnail{
		ArticleId: ad.Id,
		Processed: true,
	}

	t.logger.Debugf("Generating thumbnail for article %s\n", a)

	td.Thumbnail, td.Link =
		generateThumbnailFromDescription(strings.NewReader(ad.Description))

	if td.Link == "" {
		t.logger.Debugf("%s description doesn't contain suitable link, getting extract\n", a)

		extract := a.Extract()
		if a.HasErr() {
			return a.Err()
		}

		extractData := extract.Data()

		if extract.HasErr() {
			switch err := extract.Err(); err {
			case content.ErrNoContent:
				t.logger.Debugf("Generating article extract for %s\n", a)
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
			t.logger.Debugf("Extract for %s doesn't contain a top image\n", a)
		} else {
			t.logger.Debugf("Generating thumbnail from top image %s of %s\n", extractData.TopImage, a)
			td.Thumbnail = generateThumbnailFromImageLink(extractData.TopImage)
			td.Link = extractData.TopImage
		}
	}

	thumbnail.Data(td)
	if thumbnail.Update(); thumbnail.HasErr() {
		return fmt.Errorf("Error saving thumbnail of %s to database :%v\n", a, thumbnail.Err())
	}

	return nil
}
