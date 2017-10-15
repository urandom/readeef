package thumbnail

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/nfnt/resize"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/pool"
)

const (
	minTopImageArea = 320 * 240
)

type Generator interface {
	Generate(article content.Article) error
}

func generateThumbnail(r io.Reader) (b []byte, mimeType string, err error) {
	img, imgType, err := image.Decode(r)
	if err != nil {
		return
	}

	thumb := resize.Thumbnail(380, 285, img, resize.Lanczos3)
	buf := pool.Buffer.Get()
	defer pool.Buffer.Put(buf)

	switch imgType {
	case "gif":
		if err = gif.Encode(buf, thumb, nil); err != nil {
			return
		}

		mimeType = "image/gif"
	default:
		if err = jpeg.Encode(buf, thumb, &jpeg.Options{Quality: 80}); err != nil {
			return
		}

		mimeType = "image/jpeg"
	}

	b = buf.Bytes()
	return
}

func generateThumbnailFromDescription(description io.Reader) (string, string) {
	var data, link string
	if d, err := goquery.NewDocumentFromReader(description); err == nil {
		d.Find("img").EachWithBreak(func(i int, s *goquery.Selection) bool {
			if src, ok := s.Attr("src"); ok {
				u, err := url.Parse(src)
				if err != nil || !u.IsAbs() {
					return true
				}

				resp, err := http.Get(u.String())
				if err != nil {
					return true
				}
				defer resp.Body.Close()

				buf := pool.Buffer.Get()
				defer pool.Buffer.Put(buf)

				if _, err := buf.ReadFrom(resp.Body); err != nil {
					return true
				}

				r := bytes.NewReader(buf.Bytes())

				imgCfg, _, err := image.DecodeConfig(r)
				if err != nil {
					return true
				}

				if imgCfg.Width*imgCfg.Height > minTopImageArea {
					r.Seek(0, 0)

					b, mimeType, err := generateThumbnail(r)
					if err == nil {
						link = u.String()
						data = base64DataUri(b, mimeType)
					} else {
						return true
					}

					return false
				}

			}

			return true
		})
	}

	return data, link
}

func generateThumbnailFromImageLink(link string) (t string) {
	u, err := url.Parse(link)
	if err != nil || !u.IsAbs() {
		return
	}

	resp, err := http.Get(u.String())
	if err != nil {
		return
	}
	defer resp.Body.Close()

	buf := pool.Buffer.Get()
	defer pool.Buffer.Put(buf)

	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return
	}

	b, mimeType, err := generateThumbnail(buf)
	if err == nil {
		link = u.String()
		t = base64DataUri(b, mimeType)
	} else {
		return
	}

	return
}

func generateThumbnailProcessors(articles []content.Article) <-chan content.Article {
	processors := make(chan content.Article)

	go func() {
		defer close(processors)

		for _, a := range articles {
			processors <- a
		}
	}()

	return processors
}

func base64DataUri(b []byte, mimeType string) string {
	return "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(b)
}
