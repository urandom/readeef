package popularity

import (
	"runtime"
	"sync"
)

type scoreProvider interface {
	Score(link string) (int, error)
}

type scoreRequest struct {
	scoreProvider scoreProvider
	link          string
}

type scoreResponse struct {
	score int
	err   error
}

var scoreProviders = []scoreProvider{Facebook{}, Linkedin{}, Twitter{}, Reddit{}}

func Score(link, text string) (int, error) {
	done := make(chan struct{})
	defer close(done)

	requests := generateRequests(realLink(link, text))
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

	score := -1
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

func generateRequests(link string) <-chan scoreRequest {
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
