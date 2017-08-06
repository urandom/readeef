package ttrss

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/processor"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/content/search"
	"github.com/urandom/readeef/log"
)

type request struct {
	Op            string              `json:"op"`
	Sid           string              `json:"sid"`
	Seq           int                 `json:"seq"`
	User          string              `json:"user"`
	Password      string              `json:"password"`
	OutputMode    string              `json:"output_mode"`
	UnreadOnly    bool                `json:"unread_only"`
	IncludeEmpty  bool                `json:"include_empty"`
	Limit         int                 `json:"limit"`
	Offset        int                 `json:"offset"`
	CatId         content.TagID       `json:"cat_id"`
	FeedId        content.FeedID      `json:"feed_id"`
	Skip          int                 `json:"skip"`
	IsCat         bool                `json:"is_cat"`
	ShowContent   bool                `json:"show_content"`
	ShowExcerpt   bool                `json:"show_excerpt"`
	ViewMode      string              `json:"view_mode"`
	SinceId       content.ArticleID   `json:"since_id"`
	Sanitize      bool                `json:"sanitize"`
	HasSandbox    bool                `json:"has_sandbox"`
	IncludeHeader bool                `json:"include_header"`
	OrderBy       string              `json:"order_by"`
	Search        string              `json:"search"`
	ArticleIds    []content.ArticleID `json:"article_ids"`
	Mode          int                 `json:"mode"`
	Field         int                 `json:"field"`
	Data          string              `json:"data"`
	ArticleId     []content.ArticleID `json:"article_id"`
	PrefName      string              `json:"pref_name"`
	FeedUrl       string              `json:"feed_url"`
}

type response struct {
	Seq     int             `json:"seq"`
	Status  int             `json:"status"`
	Content json.RawMessage `json:"content"`
}

type action func(request, content.User, repo.Service) (interface{}, error)

type errorContent struct {
	Error string `json:"error"`
}

const (
	API_STATUS_OK  = 0
	API_STATUS_ERR = 1
	API_VERSION    = "1.8.0"
	API_LEVEL      = 12

	ARCHIVED_ID      = 0
	FAVORITE_ID      = -1
	PUBLISHED_ID     = -2
	FRESH_ID         = -3
	ALL_ID           = -4
	RECENTLY_READ_ID = -6

	FRESH_DURATION = -24 * time.Hour

	CAT_UNCATEGORIZED      = 0
	CAT_SPECIAL            = -1 // Starred, Published, Archived, etc.
	CAT_LABELS             = -2
	CAT_ALL_EXCEPT_VIRTUAL = -3 // i.e: labels
	CAT_ALL                = -4
)

var (
	actions = make(map[string]action)
)

func Handler(
	ctx context.Context,
	service repo.Service,
	searchProvider search.Provider,
	feedManager *readeef.FeedManager,
	processors []processor.Article,
	secret []byte,
	update time.Duration,
	log log.Log,
) http.HandlerFunc {
	sessionManager := newSession(ctx)

	processors = filterProcessors(processors)

	registerAuthActions(sessionManager, secret)
	registerArticleActions(searchProvider, processors)
	registerSettingActions(feedManager, update)

	return func(w http.ResponseWriter, r *http.Request) {
		resp := response{}

		req, stop := readJSON(w, r.Body)
		if stop {
			return
		}

		var err error
		var user content.User
		var con interface{}

		log.Debugf("Request: %#v\n", req)

		resp.Seq = req.Seq

		userRepo := service.UserRepo()
		if req.Op == "login" {
			user, err = userRepo.Get(content.Login(req.User))
		} else if req.Op != "isLoggedIn" {
			if sess := sessionManager.get(req.Sid); sess.login != "" {
				user, err = userRepo.Get(content.Login(sess.login))
				if err == nil {
					sess.lastVisit = time.Now()
					sessionManager.set(req.Sid, sess)
				}
			} else {
				err = errors.WithStack(newErr("no session", "NOT_LOGGED_IN"))
			}
		}

		if err != nil {
			err = errors.WithStack(newErr(err.Error(), "NOT_LOGGED_IN"))
		}

		if err == nil {
			log.Debugf("TT-RSS OP: %s\n", req.Op)

			a, ok := actions[req.Op]
			if !ok {
				a = unknown
			}

			con, err = a(req, user, service)
		}

		if err == nil {
			resp.Status = API_STATUS_OK
		} else {
			log.Infof("Error processing TT-RSS API request: %+v\n", err)
			resp.Status = API_STATUS_ERR
			con = errorContent{Error: errorKind(err)}
		}

		writeJson(w, req, resp, con, log)
	}
}

func FakeWebHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func readJSON(w http.ResponseWriter, r io.Reader) (req request, stop bool) {
	in := map[string]interface{}{}

	if b, err := ioutil.ReadAll(r); err == nil {
		if err = json.Unmarshal(b, &in); err != nil {
			http.Error(w, "Error decoding JSON request: "+err.Error(), http.StatusBadRequest)
			return req, true
		}
	} else {
		http.Error(w, "Error reading request body: "+err.Error(), http.StatusInternalServerError)
		return req, true
	}

	return convertRequest(in), false
}

func writeJson(w http.ResponseWriter, req request, resp response, data interface{}, log log.Log) {
	b, err := json.Marshal(data)
	if err == nil {
		resp.Content = json.RawMessage(b)
	}

	b, err = json.Marshal(&resp)

	if err == nil {
		w.Header().Set("Content-Type", "text/json")
		w.Header().Set("Api-Content-Length", strconv.Itoa(len(b)))
		w.Write(b)

		log.Debugf("Output for %s: %s\n", req.Op, string(b))
	} else {
		log.Print(fmt.Errorf("TT-RSS error %s: %v", req.Op, err))

		w.WriteHeader(http.StatusInternalServerError)
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

type err struct {
	message string
	kind    string
}

func newErr(message, kind string) err {
	return err{message: message, kind: kind}
}

func (e err) Error() string {
	return e.message
}

func (e err) Kind() string {
	return e.kind
}

func errorKind(err error) string {
	type kinder interface {
		Kind() string
	}

	if v, ok := err.(kinder); ok {
		return v.Kind()
	}

	return ""
}
