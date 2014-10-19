package popularity

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type Linkedin struct{}

type linkedinResult struct {
	Count int `json:"count"`
}

func (l Linkedin) Score(link string) (int, error) {
	var score int = -1

	link = url.QueryEscape(link)

	r, err := http.Get("http://www.linkedin.com/countserv/count/share?url=" + link + "&format=json")

	if err != nil {
		return score, err
	}
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)

	var result linkedinResult
	if err := dec.Decode(&result); err != nil {
		return score, err
	}

	score = result.Count

	return score, nil
}

func (l Linkedin) String() string {
	return "Linkedin score provider"
}