package popularity

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Reddit struct{}

type redditResult struct {
	Data struct {
		Children []struct {
			Data struct {
				Score    int64 `json:"score"`
				Comments int64 `json:"num_comments"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

func (r Reddit) Score(link string) (int64, error) {
	var score int64 = -1

	link = url.QueryEscape(link)

	resp, err := http.Get("http://buttons.reddit.com/button_info.json?url=" + link)

	if err != nil {
		return score, err
	}
	defer func() {
		// Drain the body so that the connection can be reused
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	dec := json.NewDecoder(resp.Body)

	var result redditResult
	if err := dec.Decode(&result); err != nil {
		return score, fmt.Errorf("Error scoring link %s using reddit: %v", link, err)
	}

	score = 0
	for _, d := range result.Data.Children {
		score += d.Data.Score + d.Data.Comments

	}

	return score, nil
}

func (re Reddit) String() string {
	return "Reddit score provider"
}
