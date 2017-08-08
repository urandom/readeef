package feed

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/log"
	"github.com/urandom/readeef/parser"
	"github.com/urandom/readeef/pool"
)

var (
	domainPattern  = regexp.MustCompile(`^(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`)
	commentPattern = regexp.MustCompile("<!--.*?-->")
	linkPattern    = regexp.MustCompile(`<link ([^>]+)>`)
)

func Search(query string, log log.Log) (map[string]parser.Feed, error) {
	if u, err := url.Parse(query); err == nil && (u.IsAbs() || domainPattern.MatchString(u.String())) {
		if u.Scheme == "" {
			u.Scheme = "http"
		}

		return searchByURL(u, log)
	}

	// Assume the query is not a url
	return searchByQuery(query, log)
}

func searchByURL(u *url.URL, log log.Log) (map[string]parser.Feed, error) {
	log.Infof("Searching for feeds from url %s", u)
	if u.Scheme == "http" {
		u.Scheme = "https"

		if feeds, err := downloadLinkContent(u, log); err == nil {
			return feeds, nil
		}

		u.Scheme = "http"
	}

	feeds, err := downloadLinkContent(u, log)
	if err != nil {
		return nil, errors.WithMessage(err, "searching by url "+u.String())
	}

	log.Debugf("Found %d feeds", len(feeds))

	return feeds, nil
}

func searchByQuery(query string, log log.Log) (map[string]parser.Feed, error) {
	log.Infof("Searching for feeds via %s", query)
	doc, err := goquery.NewDocument("https://www.google.com/search?q=" + url.QueryEscape(query))
	if err != nil {
		return nil, errors.Wrapf(err, "querying google with %s", query)
	}

	links := doc.Find("div.g .r a").Map(func(i int, s *goquery.Selection) string {
		href := s.AttrOr("data-href", s.AttrOr("href", ""))

		if !strings.HasPrefix(href, "/url") {
			return href
		}

		u, err := url.Parse(href)
		if err != nil {
			return ""
		}

		return u.Query().Get("q")
	})

	log.Debugf("Found links %#v for query %s", links, query)

	type out struct {
		data map[string]parser.Feed
		err  error
	}

	parsed := map[string]parser.Feed{}
	input := make(chan *url.URL, 5)
	output := make(chan out)
	parseErr := make(chan error)

	var wg sync.WaitGroup

	go func() {
		var mainErr error

		for _, link := range links {
			u, err := url.Parse(link)
			if err != nil {
				mainErr = errors.Wrapf(err, "parsing link %s", link)
				continue
			}

			input <- u
		}

		close(input)

		parseErr <- mainErr
	}()

	numProviders := 10
	wg.Add(numProviders)
	for i := 0; i < numProviders; i++ {
		go func() {
			for u := range input {
				res, err := downloadLinkContent(u, log)
				output <- out{res, err}
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(output)
	}()

	if perr := <-parseErr; perr != nil {
		log.Print(perr.Error())
	}

	var outErr error
	for out := range output {
		if out.err != nil {
			outErr = out.err
			continue
		}

		for k, v := range out.data {
			parsed[k] = v
		}
	}

	if len(parsed) == 0 && outErr != nil {
		return nil, outErr
	}

	log.Debugf("Found %d feeds", len(parsed))

	return parsed, nil
}

func downloadLinkContent(u *url.URL, log log.Log) (map[string]parser.Feed, error) {
	log.Debugf("Downloading content from %s", u)
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, errors.Wrapf(err, "getting link %s", u)
	}
	defer resp.Body.Close()

	buf := pool.Buffer.Get()
	defer pool.Buffer.Put(buf)

	buf.ReadFrom(resp.Body)

	if feed, err := parser.ParseFeed(buf.Bytes(), parser.ParseRss2, parser.ParseAtom, parser.ParseRss1); err == nil {
		return map[string]parser.Feed{u.String(): feed}, nil
	}

	html := commentPattern.ReplaceAllString(buf.String(), "")
	links := linkPattern.FindAllStringSubmatch(html, -1)

	feeds := map[string]parser.Feed{}
	for _, l := range links {
		attrs := l[1]
		if strings.Contains(attrs, `"application/rss+xml"`) || strings.Contains(attrs, `'application/rss+xml'`) {
			index := strings.Index(attrs, "href=")
			attr := attrs[index+6:]
			index = strings.IndexByte(attr, attrs[index+5])
			href := attr[:index]

			if docURL, err := url.Parse(href); err != nil {
				return nil, errors.Wrapf(err, "parsing feed href %s", href)
			} else {
				if !docURL.IsAbs() {
					docURL.Scheme = u.Scheme

					if docURL.Host == "" {
						docURL.Host = u.Host
					}

				}

				feedMap, err := downloadLinkContent(docURL, log)
				if err != nil {
					return nil, err
				}

				for k, v := range feedMap {
					feeds[k] = v
				}
			}
		}
	}

	return feeds, nil
}
