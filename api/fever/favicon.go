package fever

import (
	"encoding/base64"
	"net/http"
	"sync"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	rffeed "github.com/urandom/readeef/feed"
	"github.com/urandom/readeef/log"
)

type favicon struct {
	ID   content.FeedID `json:"id"`
	Data string         `json:"data"`
}

var (
	faviconCache   = map[content.FeedID]string{}
	faviconCacheMu sync.RWMutex
)

func favicons(
	r *http.Request,
	resp resp,
	user content.User,
	service repo.Service,
	log log.Log,
) error {
	log.Infoln("Fetching fever feeds favicons")

	var favicons []favicon

	feeds, err := service.FeedRepo().ForUser(user)
	if err != nil {
		return errors.WithMessage(err, "getting user feeds")
	}

	for _, f := range feeds {
		faviconCacheMu.RLock()
		data := faviconCache[f.ID]
		faviconCacheMu.RUnlock()

		if data != "" {
			favicons = append(favicons, favicon{f.ID, data})
			continue
		}

		log.Debugf("Getting favicon for %q", f.SiteLink)
		b, ct, err := rffeed.Favicon(f.SiteLink)
		if err != nil {
			log.Printf("Error getting favicon for %q: %v", f.SiteLink, err)
			continue
		}

		data = ct + ";base64," + base64.StdEncoding.EncodeToString(b)
		faviconCacheMu.Lock()
		faviconCache[f.ID] = data
		faviconCacheMu.Unlock()

		favicons = append(favicons, favicon{f.ID, data})
	}

	resp["favicons"] = favicons

	return nil
}

func init() {
	actions["favicons"] = favicons
}
