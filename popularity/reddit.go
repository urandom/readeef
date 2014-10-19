package popularity

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type Reddit struct{}

type redditResult struct {
	Data struct {
		Children []struct {
			Data struct {
				Score    int `json:"score"`
				Comments int `json:"num_comments"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

func (r Reddit) Score(link string) (int, error) {
	var score int = -1

	link = url.QueryEscape(link)

	resp, err := http.Get("http://buttons.reddit.com/button_info.json?url=" + link)

	if err != nil {
		return score, err
	}

	dec := json.NewDecoder(resp.Body)

	var result redditResult
	if err := dec.Decode(&result); err != nil {
		return score, err
	}

	score = 0
	for _, d := range result.Data.Children {
		score += d.Data.Score + d.Data.Comments

	}

	return score, nil
}
