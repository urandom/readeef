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
	"github.com/urandom/readeef/content/repo"
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
		case "Reddit":
			log.Infoln("Initializing Reddit popularity provider")
			if reddit, err := FromReddit(config, log); err == nil {
				scoreProviders = append(scoreProviders, reddit)
			} else {
				log.Printf("Error initializing Reddit popularity provider: %v", err)
			}
		case "Twitter":
			log.Infoln("Initializing Twitter popularity provider")
			scoreProviders = append(scoreProviders, FromTwitter(config, log))
		}
	}

	p.scoreProviders = scoreProviders

	return p
}

func (p Popularity) ScoreContent(ctx context.Context, service repo.Service) {
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
				p.scoreArticles(ctx, service)

				time.Sleep(10 * time.Minute)
			}
		}
	}()
}

func (p Popularity) scoreArticles(ctx context.Context, service repo.Service) error {
	articles, err := service.ArticleRepo().All(
		content.TimeRange(time.Now().AddDate(0, 0, -5), time.Now().Add(-15*time.Minute)),
	)

	if err != nil {
		return errors.WithMessage(err, "getting all articles up to 5 days ago")
	}

	scoresRepo := service.ScoresRepo()
	for _, a := range articles {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := p.scoreArticle(ctx, a, scoresRepo); err != nil {
				p.log.Printf("Error scoring article %s: %+v\n", a, err)
			}

			time.Sleep(p.delay)
		}
	}

	return nil
}

func (p Popularity) scoreArticle(ctx context.Context, article content.Article, repo repo.Scores) error {
	scores, err := repo.Get(article)
	if err != nil {
		if content.IsNoContent(err) {
			scores.ArticleID = article.ID
		} else {
			return errors.WithMessage(err, fmt.Sprintf("Getting article %s scores", article))
		}
	}
	score, err := p.calculateScore(ctx, article.Link, article.Description)
	if err != nil {
		return errors.WithMessage(err, "getting article popularity")
	}

	scores = calculateAgedScore(scores, score, article.Date)
	if scores.Score < 1 {
		return nil
	}
	if err = repo.Update(scores); err != nil {
		return errors.WithMessage(err, fmt.Sprintf("updating article %s scores", article))
	}

	return nil
}

func (p Popularity) calculateScore(ctx context.Context, link, text string) (int64, error) {
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

	p.log.Debugf("Popularity score of '%s' is %d", link, score)

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

func calculateAgedScore(scores content.Scores, score int64, date time.Time) content.Scores {
	age := ageInDays(date)
	switch age {
	case 0:
		scores.Score1 = score
	case 1:
		scores.Score2 = score - scores.Score1
	case 2:
		scores.Score3 = score - scores.Score1 - scores.Score2
	case 3:
		scores.Score4 = score - scores.Score1 - scores.Score2 - scores.Score3
	default:
		scores.Score5 = score - scores.Score1 - scores.Score2 - scores.Score3 - scores.Score4
	}

	score = scores.Calculate()
	penalty := float64(time.Now().Unix()-date.Unix()) / (60 * 60) * float64(age)

	if penalty > 0 {
		scores.Score = int64(float64(score) / penalty)
	} else {
		scores.Score = score
	}

	return scores
}

func ageInDays(published time.Time) int {
	now := time.Now()
	sub := now.Sub(published)
	return int(sub / (24 * time.Hour))
}
