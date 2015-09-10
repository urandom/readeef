package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/util"
)

const (
	TTRSS_API_STATUS_OK  = 0
	TTRSS_API_STATUS_ERR = 1
	TTRSS_VERSION        = "1.8.0"
	TTRSS_API_LEVEL      = 12
	TTRSS_SESSION_ID     = "TTRSS_SESSION_ID"
	TTRSS_USER_NAME      = "TTRSS_USER_NAME"
)

type TtRss struct {
	webfw.BasePatternController
}

type ttRssRequest struct {
	Op       string `json:"op"`
	Sid      string `json:"sid"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type ttRssResponse struct {
	Seq     int64 `json:"seq"`
	Status  int   `json:"status"`
	Content struct {
		Error     string      `json:"error,omitempty"`
		Level     int         `json:"level,omitempty"`
		ApiLevel  int         `json:"api_level,omitempty"`
		Version   string      `json:"version,omitempty"`
		SessionId string      `json:"session_id,omitempty"`
		Status    interface{} `json:"status,omitempty"`
	} `json:"content"`
}

func NewTtRss() TtRss {
	return TtRss{
		webfw.NewBasePatternController("/v:version/tt-rss/", webfw.MethodPost, ""),
	}
}

func (con TtRss) Handler(c context.Context) http.Handler {
	repo := readeef.GetRepo(c)
	logger := webfw.GetLogger(c)
	config := readeef.GetConfig(c)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		req := ttRssRequest{}
		dec := json.NewDecoder(r.Body)
		sess := webfw.GetSession(c, r)

		resp := ttRssResponse{}

		var err error
		var errType string
		var user content.User

		switch {
		default:
			err = r.ParseForm()
			if err != nil {
				break
			}

			seq, _ := strconv.ParseInt(r.Form.Get("seq"), 10, 64)

			resp.Seq = seq

			err = dec.Decode(&req)
			if err != nil {
				err = fmt.Errorf("Error decoding JSON request: %v", err)
				break
			}

			if req.Op != "login" && req.Op != "isLoggedIn" {
				if id, ok := sess.Get(TTRSS_SESSION_ID); ok && id != "" && id == req.Sid {
					userName, _ := sess.Get(TTRSS_USER_NAME)
					if login, ok := userName.(string); ok {
						user = repo.UserByLogin(data.Login(login))
						if repo.Err() != nil {
							errType = "NOT_LOGGED_IN"
						}
					} else {
						errType = "NOT_LOGGED_IN"
					}
				} else {
					errType = "NOT_LOGGED_IN"
				}
			}

			if errType != "" {
				break
			}

			switch req.Op {
			case "getApiLevel":
				resp.Content.Level = TTRSS_API_LEVEL
			case "getVersion":
				resp.Content.Version = TTRSS_VERSION
			case "login":
				user = repo.UserByLogin(data.Login(req.User))
				if repo.Err() != nil {
					errType = "LOGIN_ERROR"
					break
				}

				if !user.Authenticate(req.Password, []byte(config.Auth.Secret)) {
					errType = "LOGIN_ERROR"
					break
				}

				sessId := util.UUID()
				sess.Set(TTRSS_SESSION_ID, sessId)
				sess.Set(TTRSS_USER_NAME, req.User)

				resp.Content.ApiLevel = TTRSS_API_LEVEL
				resp.Content.SessionId = sessId
			case "logout":
				sess.Delete(TTRSS_SESSION_ID)
				sess.Delete(TTRSS_USER_NAME)

				resp.Content.Status = "OK"
			case "isLoggedIn":
				if id, ok := sess.Get(TTRSS_SESSION_ID); ok && id != "" && id == req.Sid {
					resp.Content.Status = true
				} else {
					resp.Content.Status = false
				}
			}
		}

		if err == nil && errType == "" {
			resp.Status = TTRSS_API_STATUS_OK
		} else {
			resp.Status = TTRSS_API_STATUS_ERR
			resp.Content.Error = errType
		}

		var b []byte
		b, err = json.Marshal(resp)

		if err == nil {
			w.Header().Set("Content-Type", "text/json")
			w.Write(b)
		} else {
			logger.Print(fmt.Errorf("TT-RSS error %s: %v", req.Op, err))

			w.WriteHeader(http.StatusInternalServerError)
		}

	})
}
