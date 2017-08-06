package popularity

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/config"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/log"
)

type Popularity struct {
	scoreProviders []scoreProvider
	delay          time.Duration
	log            log.Log
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

func New(config config.Popularity, log log.Log) Popularity {
	p := Popularity{delay: config.Converted.Delay, log: log}

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

func (p Popularity) ScoreContent(ctx context.Context, repo content.Repo) {
	if len(p.scoreProviders) == 0 {
		p.log.Infoln("No popularity providers configured")
		return
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				p.scoreFeeds(ctx, repo)

				time.Sleep(10 * time.Minute)
			}
		}
	}()
}

func (p Popularity) Score(ctx context.Context, link, text string) (int64, error) {
	innerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	requests := p.generateRequests(realLink(link, text))
	response := make(chan scoreResponse)

	var wg sync.WaitGroup
	numProviders := runtime.NumCPU()

	wg.Add(numProviders)
	for i := 0; i < numProviders; i++ {
		go func() {
			scoreContent(innerCtx, requests, response)
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

	p.log.Debugf("Popularity score of '%s' is %d\n", link, score)

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

func (p Popularity) scoreFeeds(ctx context.Context, repo content.Repo) {
	feeds := repo.AllFeeds()

	if repo.HasErr() {
		p.log.Printf("Error getting feeds: %+v", repo.Err())
		return
	}

	for _, f := range feeds {
		select {
		case <-ctx.Done():
			return
		default:
			if err := p.scoreArticles(ctx, f); err != nil {
				p.log.Printf("Error scoring feed %s articles: %+v\n", f, err)
			}
		}
	}
}

func (p Popularity) scoreArticles(ctx context.Context, feed content.Feed) error {
	articles := feed.LatestArticles()
	if feed.HasErr() {
		return errors.WithMessage(feed.Err(), fmt.Sprintf("getting latest feed articles from %s", feed))
	}

	time.Sleep(p.delay)

	for _, a := range articles {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := p.scoreArticle(ctx, a); err != nil {
				p.log.Printf("Error scoring article %s: %+v\n", a, err)
			}
		}
	}

	return nil
}

func (p Popularity) scoreArticle(ctx context.Context, article content.Article) error {
	scores := article.Scores()

	if article.HasErr() {
		err := article.Err()
		if err == content.ErrNoContent {
			scores.Data(data.ArticleScores{ArticleId: article.Data().Id})
		} else {
			return errors.WithMessage(err, "getting article scores")
		}
	}

	data := article.Data()
	score, err := p.Score(ctx, data.Link, data.Description)
	if err != nil {
		return errors.WithMessage(err, "getting article popularity")
	}

	scores.Data(calculateAgedScore(score, scores, data.Date))
	scores.Update()
	if scores.HasErr() {
		return errors.WithMessage(err, "updating article scores")
	}

	return nil
}

func scoreContent(ctx context.Context, reqs <-chan scoreRequest, res chan<- scoreResponse) {
	for req := range reqs {
		score, err := req.scoreProvider.Score(req.link)
		select {
		case res <- scoreResponse{score: score, err: err}:
		case <-ctx.Done():
			return
		}
	}

}

func calculateAgedScore(score int64, scores content.ArticleScores, date time.Time) data.ArticleScores {
	data := scores.Data()

	age := ageInDays(date)
	switch age {
	case 0:
		data.Score1 = score
	case 1:
		data.Score2 = score - data.Score1
	case 2:
		data.Score3 = score - data.Score1 - data.Score2
	case 3:
		data.Score4 = score - data.Score1 - data.Score2 - data.Score3
	default:
		data.Score5 = score - data.Score1 - data.Score2 - data.Score3 - data.Score4
	}

	score = scores.Calculate()
	penalty := float64(time.Now().Unix()-date.Unix()) / (60 * 60) * float64(age)

	if penalty > 0 {
		data.Score = int64(float64(score) / penalty)
	} else {
		data.Score = score
	}

	return data
}

func ageInDays(published time.Time) int {
	now := time.Now()
	sub := now.Sub(published)
	return int(sub / (24 * time.Hour))
}
