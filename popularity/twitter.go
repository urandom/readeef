package popularity

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type Twitter struct{}

type twitterResult struct {
	Count int `json:"count"`
}

func (t Twitter) Score(link string) (int, error) {
	var score int = -1

	link = url.QueryEscape(link)

	r, err := http.Get("http://urls.api.twitter.com/1/urls/count.json?url=" + link)

	if err != nil {
		return score, err
	}

	dec := json.NewDecoder(r.Body)

	var result twitterResult
	if err := dec.Decode(&result); err != nil {
		return score, err
	}

	score = result.Count

	return score, nil
}
