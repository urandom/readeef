package popularity

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/turnage/graw/reddit"
	"github.com/urandom/readeef/config"
	"github.com/urandom/readeef/log"
)

type Reddit struct {
	bot reddit.Bot
	log log.Log
}

func FromReddit(config config.Popularity, log log.Log) (Reddit, error) {
	if config.Reddit.ID == "" || config.Reddit.Secret == "" {
		return Reddit{}, errors.New("invalid credentials")
	}

	bot, err := reddit.NewBot(reddit.BotConfig{
		Agent: fmt.Sprintf("readeef:score_bot by /u/%s", config.Reddit.Username),
		App: reddit.App{
			ID:       config.Reddit.ID,
			Secret:   config.Reddit.Secret,
			Username: config.Reddit.Username,
			Password: config.Reddit.Password,
		},
	})
	if err != nil {
		return Reddit{}, errors.Wrap(err, "initializing reddit api")
	}
	return Reddit{bot: bot, log: log}, nil
}

func (r Reddit) Score(link string) (int64, error) {
	var score int64 = -1

	harvest, err := r.bot.ListingWithParams("/api/info", map[string]string{
		"url": link,
	})

	if err != nil {
		return score, errors.Wrapf(err, "getting reddit info for link %s", link)
	}

	score = 0
	for _, post := range harvest.Posts {
		score += int64(post.Score + post.NumComments)
	}

	r.log.Debugf("Popularity: Reddit score and comments for url %s: %d", link, score)

	return score, nil
}

func (re Reddit) String() string {
	return "Reddit score provider"
}
