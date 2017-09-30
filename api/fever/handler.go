package fever

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/processor"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type resp map[string]interface{}

type action func(*http.Request, resp, content.User, repo.Service, log.Log) error

const (
	API_VERSION = 2
)

var (
	actions = make(map[string]action)
)

func Handler(
	service repo.Service,
	processors []processor.Article,
	log log.Log,
) http.HandlerFunc {

	processors = filterProcessors(processors)

	registerItemActions(processors)
	registerLinkActions(processors)

	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var user content.User

		err = r.ParseForm()

		if err == nil {
			user, err = readeefUser(service.UserRepo(), r.FormValue("api_key"), log)
		}

		resp := resp{"api_version": API_VERSION}

		if user.Login == "" || err != nil {
			if err != nil {
				log.Printf("Error getting readeef user: %+v\n", err)
			}
			resp["auth"] = 0

			writeJSON(w, resp)
			return
		}

		now := time.Now().Unix()

		resp["auth"] = 1
		resp["last_refreshed_on_time"] = now

		var lastName string

		for name, action := range actions {
			if _, ok := r.Form[name]; ok {
				if err = action(r, resp, user, service, log); err != nil {
					lastName = name
					break
				}
			}
		}

		if err != nil {
			log.Printf("Error calling %s: %+v", lastName, err)
		}

		writeJSON(w, resp)
	}
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	if b, err := json.Marshal(data); err == nil {
		w.Write(b)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func filterProcessors(input []processor.Article) []processor.Article {
	processors := make([]processor.Article, 0, len(input))

	for i := range input {
		if _, ok := input[i].(processor.ProxyHTTP); ok {
			continue
		}

		processors = append(processors, input[i])
	}

	return processors
}
