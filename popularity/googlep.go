package popularity

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/urandom/webfw/util"
)

type GoogleP struct{}

type googlepResult struct {
	Result struct {
		Metadata struct {
			GlobalCounts struct {
				Count float64 `json:"count"`
			} `json:"globalCounts"`
		} `json:"metadata"`
	} `json:"result"`
}

func (f GoogleP) Score(link string) (int64, error) {
	var score int64 = -1

	buf := util.BufferPool.GetBuffer()
	defer util.BufferPool.Put(buf)

	request := map[string]interface{}{
		"method":     "pos.plusones.get",
		"id":         "p",
		"jsonrpc":    "2.0",
		"key":        "p",
		"apiVersion": "v1",
	}
	request["params"] = map[string]interface{}{
		"nolog":   true,
		"id":      link,
		"source":  "widget",
		"userId":  "@viewer",
		"groupId": "@self",
	}

	requestData, err := json.Marshal([]interface{}{request})
	if err != nil {
		return score, fmt.Errorf("Error marshaling google score data for %s: %v", link, err)
	}

	_, err = buf.Write(requestData)
	if err != nil {
		return score, err
	}

	r, err := http.Post("https://clients6.google.com/rpc", "application/json", buf)

	if err != nil {
		return score, err
	}
	defer func() {
		// Drain the body so that the connection can be reused
		io.Copy(ioutil.Discard, r.Body)
		r.Body.Close()
	}()

	dec := json.NewDecoder(r.Body)

	var results []googlepResult
	if err := dec.Decode(&results); err != nil {
		return score, fmt.Errorf("Error scoring link %s using google: %v", link, err)
	}

	score = 0
	for _, d := range results {
		score += int64(d.Result.Metadata.GlobalCounts.Count)
	}

	return score, nil
}

func (f GoogleP) String() string {
	return "GoogleP score provider"
}
