package popularity

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Facebook struct{}

type facebookResult struct {
	Shares   int64 `json:"share_count"`
	Likes    int64 `json:"like_count"`
	Comments int64 `json:"comment_count"`
}

func (f Facebook) Score(link string) (int64, error) {
	var score int64 = -1

	link = url.QueryEscape(link)

	r, err := http.Get("https://api.facebook.com/method/links.getStats?urls=" + link + "&format=json")

	if err != nil {
		return score, err
	}
	defer func() {
		// Drain the body so that the connection can be reused
		io.Copy(ioutil.Discard, r.Body)
		r.Body.Close()
	}()

	dec := json.NewDecoder(r.Body)

	var results []facebookResult
	if err := dec.Decode(&results); err != nil {
		return score, fmt.Errorf("Error scoring link %s using facebook: %v", link, err)
	}

	score = 0
	for _, d := range results {
		score += int64(float64(d.Likes)*0.01) + d.Shares + d.Comments

	}

	return score, nil
}

func (f Facebook) String() string {
	return "Facebook score provider"
}
