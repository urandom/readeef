package extractor

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

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
		return nil, fmt.Errorf("Readability API key cannot be empty")
	}
	return Readability{key: key}, nil
}

func (e Readability) Extract(link string) (data data.ArticleExtract, err error) {
	url := fmt.Sprintf("http://readability.com/api/content/v1/parser?url=%s&token=%s",
		url.QueryEscape(link), e.key,
	)

	var r readability
	var resp *http.Response

	resp, err = http.Get(url)

	if err != nil {
		return
	}

	defer func() {
		// Drain the body so that the connection can be reused
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()
	dec := json.NewDecoder(resp.Body)

	err = dec.Decode(&r)
	if err != nil {
		err = fmt.Errorf("Error extracting content from %s: %v", link, err)
		return
	}

	data.Title = r.Title
	data.Content = r.Content
	data.TopImage = r.LeadImage
	data.Language = "en"
	return
}
