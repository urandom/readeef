package extract

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
)

type readability struct {
	key string
}

type readabilityData struct {
	Content   string
	Title     string
	LeadImage string `json:"lead_image_url"`
}

func WithReadability(key string) (content.Extractor, error) {
	if key == "" {
		return nil, errors.New("Readability API key cannot be empty")
	}
	return readability{key: key}, nil
}

func (e readability) Extract(link string) (content.ArticleExtract, error) {
	url := fmt.Sprintf("http://readability.com/api/content/v1/parser?url=%s&token=%s",
		url.QueryEscape(link), e.key,
	)

	var r readabilityData

	resp, err := http.Get(url)

	if err != nil {
		return content.ArticleExtract{}, errors.Wrap(err, "getting url response")
	}

	defer func() {
		// Drain the body so that the connection can be reused
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()
	dec := json.NewDecoder(resp.Body)

	err = dec.Decode(&r)
	if err != nil {
		return content.ArticleExtract{}, errors.Wrapf(err, "extracting content from %s", link)
	}

	data := content.ArticleExtract{}
	data.Title = r.Title
	data.Content = r.Content
	data.TopImage = r.LeadImage
	data.Language = "en"

	return data, nil
}
