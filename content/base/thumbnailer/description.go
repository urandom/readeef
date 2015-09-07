package thumbnailer

import (
	"fmt"
	"strings"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/webfw"
)

type Description struct {
	logger webfw.Logger
}

func NewDescription(l webfw.Logger) content.Thumbnailer {
	return Description{logger: l}
}

func (t Description) Generate(a content.Article) error {
	ad := a.Data()

	thumbnail := a.Repo().ArticleThumbnail()
	td := data.ArticleThumbnail{
		ArticleId: ad.Id,
		Processed: true,
	}

	t.logger.Debugf("Generating thumbnail for article %s\n", a)

	td.Thumbnail, td.Link =
		generateThumbnailFromDescription(strings.NewReader(ad.Description))

	thumbnail.Data(td)
	if thumbnail.Update(); thumbnail.HasErr() {
		return fmt.Errorf("Error saving thumbnail of %s to database :%v\n", a, thumbnail.Err())
	}

	return nil
}
