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

func WithReadability(key string) (Generator, error) {
	if key == "" {
		return nil, errors.New("Readability API key cannot be empty")
	}
	return readability{key: key}, nil
}

func (e readability) Generate(link string) (content.Extract, error) {
	url := fmt.Sprintf("http://readability.com/api/content/v1/parser?url=%s&token=%s",
		url.QueryEscape(link), e.key,
	)

	var r readabilityData

	resp, err := http.Get(url)

	if err != nil {
		return content.Extract{}, errors.Wrap(err, "getting url response")
	}

	defer func() {
		// Drain the body so that the connection can be reused
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()
	dec := json.NewDecoder(resp.Body)

	err = dec.Decode(&r)
	if err != nil {
		return content.Extract{}, errors.Wrapf(err, "extracting content from %s", link)
	}

	extract := content.Extract{}
	extract.Title = r.Title
	extract.Content = r.Content
	extract.TopImage = r.LeadImage
	extract.Language = "en"

	return extract, nil
}
