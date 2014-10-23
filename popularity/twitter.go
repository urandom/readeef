package popularity

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type Twitter struct{}

type twitterResult struct {
	Count int64 `json:"count"`
}

func (t Twitter) Score(link string) (int64, error) {
	var score int64 = -1

	link = url.QueryEscape(link)

	r, err := http.Get("http://urls.api.twitter.com/1/urls/count.json?url=" + link)

	if err != nil {
		return score, err
	}
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)

	var result twitterResult
	if err := dec.Decode(&result); err != nil {
		return score, err
	}

	score = result.Count

	return score, nil
}

func (t Twitter) String() string {
	return "Twitter score provider"
}
