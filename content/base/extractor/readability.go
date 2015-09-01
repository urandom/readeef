package extractor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

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

func NewReadability(key string) (Readability, error) {
	if key == "" {
		return Readability{}, fmt.Errorf("Readability API key cannot be empty")
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

	defer resp.Body.Close()
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
