package thumbnailer

import (
	"fmt"
	_ "image/png"
	"strings"
	"sync"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/webfw"
)

type Description struct {
	logger webfw.Logger
}

func NewDescription(l webfw.Logger) Description {
	return Description{logger: l}
}

func (t Description) Process(articles []content.Article) error {
	t.logger.Debugln("Generating thumbnailer processors")

	processors := generateThumbnailProcessors(articles)
	numProcessors := 20
	done := make(chan struct{})
	errc := make(chan error)

	defer close(done)

	var wg sync.WaitGroup

	wg.Add(numProcessors)
	for i := 0; i < numProcessors; i++ {
		go func() {
			err := t.processThumbnail(done, processors)
			if err != nil {
				errc <- err
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(errc)
	}()

	for err := range errc {
		return err
	}

	return nil
}

func (t Description) processThumbnail(done <-chan struct{}, processors <-chan content.Article) error {
	for a := range processors {
		select {
		case <-done:
			return nil
		default:
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
		}
	}

	return nil
}
