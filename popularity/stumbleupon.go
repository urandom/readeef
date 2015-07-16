package popularity

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type StumbleUpon struct{}

type StumbleUponResult struct {
	Result struct {
		Views int64 `json:"views"`
	} `json:"result"`
}

func (t StumbleUpon) Score(link string) (int64, error) {
	var score int64 = -1

	link = url.QueryEscape(link)

	r, err := http.Get("http://www.stumbleupon.com/services/1.01/badge.getinfo?url=" + link)

	if err != nil {
		return score, err
	}
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)

	var result StumbleUponResult
	if err := dec.Decode(&result); err != nil {
		return score, fmt.Errorf("Error scoring link %s using stumbleupon: %v", link, err)
	}

	score = result.Result.Views

	return score, nil
}

func (t StumbleUpon) String() string {
	return "StumbleUpon score provider"
}
