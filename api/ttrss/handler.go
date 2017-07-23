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
	"github.com/urandom/handler/method"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

type request struct {
	Op            string           `json:"op"`
	Sid           string           `json:"sid"`
	Seq           int              `json:"seq"`
	User          string           `json:"user"`
	Password      string           `json:"password"`
	OutputMode    string           `json:"output_mode"`
	UnreadOnly    bool             `json:"unread_only"`
	IncludeEmpty  bool             `json:"include_empty"`
	Limit         int              `json:"limit"`
	Offset        int              `json:"offset"`
	CatId         data.TagId       `json:"cat_id"`
	FeedId        data.FeedId      `json:"feed_id"`
	Skip          int              `json:"skip"`
	IsCat         bool             `json:"is_cat"`
	ShowContent   bool             `json:"show_content"`
	ShowExcerpt   bool             `json:"show_excerpt"`
	ViewMode      string           `json:"view_mode"`
	SinceId       data.ArticleId   `json:"since_id"`
	Sanitize      bool             `json:"sanitize"`
	HasSandbox    bool             `json:"has_sandbox"`
	IncludeHeader bool             `json:"include_header"`
	OrderBy       string           `json:"order_by"`
	Search        string           `json:"search"`
	ArticleIds    []data.ArticleId `json:"article_ids"`
	Mode          int              `json:"mode"`
	Field         int              `json:"field"`
	Data          string           `json:"data"`
	ArticleId     []data.ArticleId `json:"article_id"`
	PrefName      string           `json:"pref_name"`
	FeedUrl       string           `json:"feed_url"`
}

type response struct {
	Seq     int             `json:"seq"`
	Status  int             `json:"status"`
	Content json.RawMessage `json:"content"`
}

type action func(request, content.User) (interface{}, error)

type errorContent struct {
	Error string `json:"error"`
}

const (
	TTRSS_API_STATUS_OK  = 0
	TTRSS_API_STATUS_ERR = 1
	TTRSS_VERSION        = "1.8.0"
	TTRSS_API_LEVEL      = 12

	TTRSS_ARCHIVED_ID      = 0
	TTRSS_FAVORITE_ID      = -1
	TTRSS_PUBLISHED_ID     = -2
	TTRSS_FRESH_ID         = -3
	TTRSS_ALL_ID           = -4
	TTRSS_RECENTLY_READ_ID = -6

	TTRSS_FRESH_DURATION = -24 * time.Hour

	TTRSS_CAT_UNCATEGORIZED      = 0
	TTRSS_CAT_SPECIAL            = -1 // Starred, Published, Archived, etc.
	TTRSS_CAT_LABELS             = -2
	TTRSS_CAT_ALL_EXCEPT_VIRTUAL = -3 // i.e: labels
	TTRSS_CAT_ALL                = -4
)

var (
	actions = make(map[string]action)
)

// /v{TTRSS_API_LEVEL}/tt-rss/api/
func Handler(
	ctx context.Context,
	repo content.Repo,
	searchProvider content.SearchProvider,
	feedManager readeef.FeedManager,
	secret []byte,
	update time.Duration,
	log readeef.Logger,
) http.Handler {
	sessionManager := newSession(ctx)

	registerAuthActions(sessionManager, secret)
	registerArticleActions(searchProvider)
	registerSettingActions(feedManager, update)

	return method.Post(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		if req.Op == "login" {
			user = repo.UserByLogin(data.Login(req.User))
		} else if req.Op != "isLoggedIn" {
			if sess := sessionManager.get(req.Sid); sess.login != "" {
				user = repo.UserByLogin(data.Login(sess.login))
				if !user.HasErr() {
					sess.lastVisit = time.Now()
					sessionManager.set(req.Sid, sess)
				}
			} else {
				err = errors.WithStack(newErr("no session", "NOT_LOGGED_IN"))
			}
		}

		if user != nil && user.HasErr() {
			err = errors.WithStack(newErr(user.Err().Error(), "NOT_LOGGED_IN"))
		}

		if err == nil {
			log.Debugf("TT-RSS OP: %s\n", req.Op)

			a, ok := actions[req.Op]
			if !ok {
				a = unknown
			}

			con, err = a(req, user)
		}

		if err == nil {
			resp.Status = TTRSS_API_STATUS_OK
		} else {
			log.Infof("Error processing TT-RSS API request: %+v\n", err)
			resp.Status = TTRSS_API_STATUS_ERR
			con = errorContent{Error: errorKind(err)}
		}

		writeJson(w, req, resp, con, log)
	}))
}

// /v{TTRSS_API_LEVEL}/tt-rss
func FakeWebHandler() http.Handler {
	return method.Get(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	}))
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

func writeJson(w http.ResponseWriter, req request, resp response, data interface{}, log readeef.Logger) {
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
