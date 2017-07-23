package fever

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
)

type resp map[string]interface{}

type action struct {
	name string
	call func(*http.Request, resp, content.User, readeef.Logger) error
}

const (
	FEVER_API_VERSION = 2
)

var (
	actions = [...]action{
		{"groups", groups},
		{"feeds", feeds},
		{"unread_item_ids", unreadItemIDs},
		{"saved_item_ids", savedItemIDs},
		{"items", items},
		{"links", links},
		{"unread_recently_read", unreadRecent},
		{"mark", markItem},
	}
)

// /api/v2/fever/
func Handler(repo content.Repo, log readeef.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		var user content.User

		err = r.ParseForm()

		if err == nil {
			user = readeefUser(repo, r.FormValue("api_key"), log)
		}

		resp := resp{"api_version": FEVER_API_VERSION}

		if user == nil || user.HasErr() {
			resp["auth"] = 0

			writeJSON(w, resp)
			return
		}

		now := time.Now().Unix()

		resp["auth"] = 1
		resp["last_refreshed_on_time"] = now

		var lastName string

		for _, action := range actions {
			if _, ok := r.Form[action.name]; ok {
				if err = action.call(r, resp, user, log); err != nil {
					lastName = action.name
					break
				}
			}
		}

		if err != nil {
			log.Print("Error calling %s: %+v", lastName, err)
		}

		writeJSON(w, resp)
	})
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	if b, err := json.Marshal(data); err == nil {
		w.Write(b)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
