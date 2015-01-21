package popularity

import (
	"runtime"
	"sync"
)

type scoreProvider interface {
	Score(link string) (int64, error)
}

type scoreRequest struct {
	scoreProvider scoreProvider
	link          string
}

type scoreResponse struct {
	score int64
	err   error
}

func Score(link, text string, providers []string) (int64, error) {
	done := make(chan struct{})
	defer close(done)

	scoreProviders := []scoreProvider{}
	for _, p := range providers {
		switch p {
		case "Facebook":
			scoreProviders = append(scoreProviders, Facebook{})
		case "GoogleP":
			scoreProviders = append(scoreProviders, GoogleP{})
		case "Linkedin":
			scoreProviders = append(scoreProviders, Linkedin{})
		case "Reddint":
			scoreProviders = append(scoreProviders, Reddit{})
		case "StumbleUpon":
			scoreProviders = append(scoreProviders, StumbleUpon{})
		case "Twitter":
			scoreProviders = append(scoreProviders, Twitter{})
		}
	}

	requests := generateRequests(realLink(link, text), scoreProviders)
	response := make(chan scoreResponse)

	var wg sync.WaitGroup
	numProviders := runtime.NumCPU()

	wg.Add(numProviders)
	for i := 0; i < numProviders; i++ {
		go func() {
			scoreContent(done, requests, response)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(response)
	}()

	score := int64(-1)
	var err error
	for resp := range response {
		if resp.err == nil {
			score += resp.score
		} else {
			err = resp.err
		}
	}

	if score == -1 && err != nil {
		return score, err
	}

	return score + 1, nil
}

func generateRequests(link string, scoreProviders []scoreProvider) <-chan scoreRequest {
	providers := make(chan scoreRequest)

	go func() {
		defer close(providers)

		for _, p := range scoreProviders {
			providers <- scoreRequest{scoreProvider: p, link: link}
		}
	}()

	return providers
}

func scoreContent(done <-chan struct{}, reqs <-chan scoreRequest, res chan<- scoreResponse) {
	for req := range reqs {
		score, err := req.scoreProvider.Score(req.link)
		select {
		case res <- scoreResponse{score: score, err: err}:
		case <-done:
			return
		}
	}

}
