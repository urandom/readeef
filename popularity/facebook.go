package popularity

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type Facebook struct{}

type facebookResult struct {
	Shares   int `json:"share_count"`
	Likes    int `json:"like_count"`
	Comments int `json:"comment_count"`
}

func (f Facebook) Score(link string) (int, error) {
	var score int = -1

	link = url.QueryEscape(link)

	r, err := http.Get("https://api.facebook.com/method/links.getStats?urls=" + link + "&format=json")

	if err != nil {
		return score, err
	}

	dec := json.NewDecoder(r.Body)

	var results []facebookResult
	if err := dec.Decode(&results); err != nil {
		return score, err
	}

	score = 0
	for _, d := range results {
		score += d.Likes + d.Shares + d.Comments

	}

	return score, nil
}
