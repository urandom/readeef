package popularity

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Linkedin struct{}

type linkedinResult struct {
	Count int64 `json:"count"`
}

func (l Linkedin) Score(link string) (int64, error) {
	var score int64 = -1

	link = url.QueryEscape(link)

	r, err := http.Get("http://www.linkedin.com/countserv/count/share?url=" + link + "&format=json")

	if err != nil {
		return score, err
	}
	defer func() {
		// Drain the body so that the connection can be reused
		io.Copy(ioutil.Discard, r.Body)
		r.Body.Close()
	}()

	dec := json.NewDecoder(r.Body)

	var result linkedinResult
	if err := dec.Decode(&result); err != nil {
		return score, fmt.Errorf("Error scoring link %s using linkedin: %v", link, err)
	}

	score = result.Count

	return score, nil
}

func (l Linkedin) String() string {
	return "Linkedin score provider"
}
