package popularity

import (
	"fmt"
	"net/url"

	"github.com/ChimeraCoder/anaconda"
	"github.com/urandom/readeef/config"
)

type Twitter struct {
	api *anaconda.TwitterApi
}

func NewTwitter(config config.Popularity) Twitter {
	anaconda.SetConsumerKey(config.Twitter.ConsumerKey)
	anaconda.SetConsumerSecret(config.Twitter.ConsumerSecret)

	return Twitter{api: anaconda.NewTwitterApi(config.Twitter.AccessToken, config.Twitter.AccessTokenSecret)}
}

func (t Twitter) Score(link string) (int64, error) {
	link = url.QueryEscape(link)

	var score int64

	values := make(url.Values)
	values.Set("count", "100")
	for {
		searchResults, err := t.api.GetSearch(link, values)
		if err != nil {
			return 0, fmt.Errorf("twitter scoring: %v", err)
		}

		score += int64(len(searchResults.Statuses))

		if searchResults.Metadata.NextResults == "" {
			break
		}

		v, err := url.ParseQuery(searchResults.Metadata.NextResults[1:])
		if err != nil {
			panic(err)
		}

		values.Set("max_id", v.Get("max_id"))
	}

	return score, nil
}

func (t Twitter) String() string {
	return "Twitter score provider"
}
