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

type Extract struct {
	extractor content.Extractor
	logger    webfw.Logger
}

func NewExtract(e content.Extractor, l webfw.Logger) (Extract, error) {
	if e == nil {
		return Extract{}, fmt.Errorf("A valid extractor is required")
	}
	return Extract{extractor: e, logger: l}, nil
}

func (t Extract) Process(articles []content.Article) error {
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

func (t Extract) processThumbnail(done <-chan struct{}, processors <-chan content.Article) error {
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

			td.Thumbnail, td.MimeType, td.Link =
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
					td.Thumbnail, td.MimeType = generateThumbnailFromImageLink(extractData.TopImage)
					td.Link = extractData.TopImage
				}
			}

			thumbnail.Data(td)
			if thumbnail.Update(); thumbnail.HasErr() {
				return fmt.Errorf("Error saving thumbnail of %s to database :%v\n", a, thumbnail.Err())
			}
		}
	}

	return nil
}
