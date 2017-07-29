package extractor

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

type Readability struct {
	key string
}

type readability struct {
	Content   string
	Title     string
	LeadImage string `json:"lead_image_url"`
}

func NewReadability(key string) (content.Extractor, error) {
	if key == "" {
		return nil, errors.New("Readability API key cannot be empty")
	}
	return Readability{key: key}, nil
}

func (e Readability) Extract(link string) (data.ArticleExtract, error) {
	url := fmt.Sprintf("http://readability.com/api/content/v1/parser?url=%s&token=%s",
		url.QueryEscape(link), e.key,
	)

	var r readability

	resp, err := http.Get(url)

	if err != nil {
		return data.ArticleExtract{}, errors.Wrap(err, "getting url response")
	}

	defer func() {
		// Drain the body so that the connection can be reused
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()
	dec := json.NewDecoder(resp.Body)

	err = dec.Decode(&r)
	if err != nil {
		return data.ArticleExtract{}, errors.Wrapf(err, "extracting content from %s", link)
	}

	data := data.ArticleExtract{}
	data.Title = r.Title
	data.Content = r.Content
	data.TopImage = r.LeadImage
	data.Language = "en"

	return data, nil
}
