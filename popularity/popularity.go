package popularity

import (
	"runtime"
	"sync"
	"time"

	"github.com/urandom/webfw"
)

type Popularity struct {
	scoreProviders []scoreProvider
	logger         webfw.Logger
}

type Config struct {
	Delay     string
	Providers []string

	TwitterConsumerKey       string `gcfg:"twitter-consumer-key"`
	TwitterConsumerSecret    string `gcfg:"twitter-consumer-secret"`
	TwitterAccessToken       string `gcfg:"twitter-access-token"`
	TwitterAccessTokenSecret string `gcfg:"twitter-access-token-secret"`

	Converted struct {
		Delay time.Duration
	}
}

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

func New(config Config, logger webfw.Logger) Popularity {
	p := Popularity{logger: logger}

	scoreProviders := []scoreProvider{}
	for _, p := range config.Providers {
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
			scoreProviders = append(scoreProviders, NewTwitter(config))
		}
	}

	p.scoreProviders = scoreProviders

	return p
}

func (p Popularity) Score(link, text string) (int64, error) {
	done := make(chan struct{})
	defer close(done)

	requests := p.generateRequests(realLink(link, text))
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

	score++

	p.logger.Debugf("Popularity score of '%s' is %d\n", link, score)

	return score, nil
}

func (p Popularity) generateRequests(link string) <-chan scoreRequest {
	providers := make(chan scoreRequest)

	go func() {
		defer close(providers)

		for _, p := range p.scoreProviders {
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
