package feed

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

func Favicon(site string) ([]byte, string, error) {
	doc, err := goquery.NewDocument(site)
	if err != nil {
		return nil, "", errors.Wrapf(err, "querying site: %q", site)
	}

	icons := doc.Find(`link[rel="shortcut icon"], link[rel="icon"], link[rel="icon shortcut"]`).
		Filter(`[href]`).
		Map(func(_ int, s *goquery.Selection) string {
			return s.AttrOr("href", "")
		})

	var siteURL *url.URL
	for _, icon := range icons {
		if icon == "" {
			continue
		}

		iconURL, err := url.Parse(icon)
		if err != nil {
			continue
		}

		if !iconURL.IsAbs() {
			if siteURL == nil {
				siteURL, err = url.Parse(site)
				if err != nil {
					continue
				}
			}

			iconURL.Scheme = siteURL.Scheme
			if iconURL.Host == "" {
				iconURL.Host = siteURL.Host
			}
		}

		resp, err := http.Get(iconURL.String())
		if err != nil {
			return nil, "", errors.Wrapf(err, "getting favicon %q", iconURL)
		}
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)

		return b, resp.Header.Get("content-type"), err
	}

	return nil, "", nil
}
