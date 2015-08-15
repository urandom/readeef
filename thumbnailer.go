package readeef

import (
	"bytes"
	"image"
	"image/gif"
	"image/jpeg"
	_ "image/png"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/nfnt/resize"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/util"
)

const (
	minTopImageArea = 320 * 240
)

type Thumbnailer struct {
	logger webfw.Logger
}

func NewThumbnailer(c Config, l webfw.Logger) Thumbnailer {
	return Thumbnailer{logger: l}
}

func (t Thumbnailer) Process(articles []content.Article) {
	t.logger.Debugln("Generating thumbnailer processors")

	processors := t.generateThumbnailProcessors(articles)
	numProcessors := 20
	done := make(chan struct{})

	defer close(done)

	var wg sync.WaitGroup

	wg.Add(numProcessors)
	for i := 0; i < numProcessors; i++ {
		go func() {
			t.processThumbnail(done, processors)
			wg.Done()
		}()
	}

	wg.Wait()

}

func (t Thumbnailer) processThumbnail(done <-chan struct{}, processors <-chan content.Article) {
	for a := range processors {
		select {
		case <-done:
			return
		default:
			ad := a.Data()

			thumbnail := a.Repo().ArticleThumbnail()
			td := data.ArticleThumbnail{
				ArticleId: ad.Id,
				Processed: true,
			}

			t.logger.Debugf("Finding suitable thumbnail images for article %s\n", a)
			if d, err := goquery.NewDocumentFromReader(strings.NewReader(ad.Description)); err == nil {
				d.Find("img").EachWithBreak(func(i int, s *goquery.Selection) bool {
					if src, ok := s.Attr("src"); ok {
						u, err := url.Parse(src)
						if err != nil || !u.IsAbs() {
							return true
						}

						t.logger.Debugf("Fetching original image of %s from %s\n", a, u)
						resp, err := http.Get(u.String())
						if err != nil {
							return true
						}
						defer resp.Body.Close()

						buf := util.BufferPool.GetBuffer()
						defer util.BufferPool.Put(buf)

						if _, err := buf.ReadFrom(resp.Body); err != nil {
							return true
						}

						r := bytes.NewReader(buf.Bytes())

						t.logger.Debugf("Decoding original image config of %s from %s\n", a, u)
						imgCfg, _, err := image.DecodeConfig(r)
						if err != nil {
							return true
						}

						t.logger.Debugf("Checking if image [%dx%d] of %s from %s is of a suitable size\n", imgCfg.Width, imgCfg.Height, a, u)
						if imgCfg.Width*imgCfg.Height > minTopImageArea {
							t.logger.Debugf("Decoding original image of %s from %s\n", a, u)
							r.Seek(0, 0)
							img, imgType, err := image.Decode(r)
							if err != nil {
								t.logger.Debugf("Error decoding image of %s from %s: %v\n", a, u, err)
								return true
							}
							td.MimeType = "image/" + imgType

							t.logger.Debugf("Generating thumbnail of %s from %s\n", a, u)
							thumb := resize.Thumbnail(192, 128, img, resize.Lanczos3)
							buf.Reset()

							switch imgType {
							case "gif":
								if err = gif.Encode(buf, thumb, nil); err != nil {
									return true
								}

								td.MimeType = "image/gif"
							default:
								if err = jpeg.Encode(buf, thumb, &jpeg.Options{Quality: 80}); err != nil {
									return true
								}

								td.MimeType = "image/jpeg"
							}

							t.logger.Debugf("Encoding thumbnail of %s from %s to type %s\n", a, u, td.MimeType)
							td.Thumbnail = buf.Bytes()

							return false
						}

					}

					return true
				})
			}

			thumbnail.Data(td)
			if thumbnail.Update(); thumbnail.HasErr() {
				t.logger.Debugf("Error saving thumbnail of %s to database :%v\n", a, thumbnail.Err())
			}
		}
	}
}

func (t Thumbnailer) generateThumbnailProcessors(articles []content.Article) <-chan content.Article {
	processors := make(chan content.Article)

	go func() {
		defer close(processors)

		for _, a := range articles {
			processors <- a
		}
	}()

	return processors
}
