package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"

	"github.com/urandom/readeef"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"golang.org/x/net/websocket"
)

type WebSocket struct {
	webfw.BasePatternController
	fm         *readeef.FeedManager
	si         readeef.SearchIndex
	updateFeed chan readeef.Feed
}

type apiRequest struct {
	Method    string          `json:"method"`
	Tag       string          `json:"tag"`
	Arguments json.RawMessage `json:"arguments"`
}

type apiResponse struct {
	Success   bool                   `json:"success"`
	ErrorType string                 `json:"errorType"`
	Error     string                 `json:"error"`
	Method    string                 `json:"method"`
	Tag       string                 `json:"tag"`
	Arguments map[string]interface{} `json:"arguments"`
}

type Processor interface {
	Process() responseError
}

var (
	errTypeInternal           = "error-internal"
	errTypeMessageParse       = "error-message-parse"
	errTypeNoId               = "error-no-id"
	errTypeInvalidMethodValue = "error-invalid-method-value"
	errTypeInvalidArgValue    = "error-invalid-arg-value"
	errTypeUnauthorized       = "error-unauthorized"

	errNoId               = errors.New("No Id given")
	errInvalidMethodValue = errors.New("Invalid method")
	errInvalidArgValue    = errors.New("Invalid argument value")
	errInternal           = errors.New("Internal server error")
	errUnauthorized       = errors.New("Unauthorized")
)

func NewWebSocket(fm *readeef.FeedManager, si readeef.SearchIndex) WebSocket {
	return WebSocket{
		BasePatternController: webfw.NewBasePatternController("/v:version/", webfw.MethodGet, ""),
		fm:         fm,
		si:         si,
		updateFeed: make(chan readeef.Feed),
	}
}

func (con WebSocket) UpdateFeedChannel() chan<- readeef.Feed {
	return con.updateFeed
}

func (con WebSocket) Handler(c context.Context) http.Handler {
	var mutex sync.RWMutex

	receivers := make(map[chan readeef.Feed]bool)
	logger := webfw.GetLogger(c)

	go func() {
		for {
			select {
			case feed := <-con.updateFeed:
				mutex.RLock()

				for receiver, _ := range receivers {
					receiver <- feed
				}

				mutex.RUnlock()
			}
		}
	}()

	return websocket.Handler(func(ws *websocket.Conn) {
		db := readeef.GetDB(c)
		user := readeef.GetUser(c, ws.Request())

		msg := make(chan apiRequest)
		resp := make(chan apiResponse)

		done := make(chan bool)
		defer close(done)

		receiver := make(chan readeef.Feed)

		mutex.Lock()
		receivers[receiver] = true
		mutex.Unlock()
		defer func() {
			mutex.Lock()
			close(receiver)
			delete(receivers, receiver)
			mutex.Unlock()
		}()

		go func() {
			for {
				var r responseError

				select {
				case data := <-msg:
					var err error
					var processor Processor

					if processor, err = data.processor(c, db, user, con.fm, con.si); err == nil {
						if len(data.Arguments) > 0 {
							err = json.Unmarshal([]byte(data.Arguments), processor)
						}

						if err == nil {
							r = processor.Process()
						}
					}

					if err != nil {
						r.err = err
						switch err.(type) {
						case *json.UnmarshalTypeError:
							r.errType = errTypeInvalidArgValue
						default:
							if err == errInvalidMethodValue {
								r.errType = errTypeInvalidMethodValue
							}
						}
					}

					go func() {
						var err string
						if r.err != nil {
							err = r.err.Error()
						}
						resp <- apiResponse{
							Success: r.err == nil, Error: err, ErrorType: r.errType,
							Method: data.Method, Tag: data.Tag, Arguments: r.val,
						}
					}()
				case f := <-receiver:
					logger.Infoln("Received notification for feed update of" + f.Link)

					r := newResponse()

					uf, _ := db.GetUserFeed(f.Id, user)

					if uf.Id > 0 {
						r.val["Feed"] = feed{
							Id: f.Id, Title: f.Title, Description: f.Description,
							Link: f.Link, Image: f.Image,
						}

						go func() {
							var err string
							if r.err != nil {
								err = r.err.Error()
							}
							resp <- apiResponse{
								Success: r.err == nil, Error: err, ErrorType: r.errType,
								Method: "feed-update-notifier", Tag: "", Arguments: r.val,
							}
						}()
					}
				case r := <-resp:
					websocket.JSON.Send(ws, r)
				case <-done:
					return
				}
			}
		}()

		for {
			var data apiRequest
			if err := websocket.JSON.Receive(ws, &data); err != nil {
				if err == io.EOF {
					// Websocket was closed
					break
				} else {
					websocket.JSON.Send(ws, apiResponse{
						Success: false, ErrorType: errTypeMessageParse,
						Error: err.Error(), Method: data.Method,
					})
				}
			}

			if forbidden(c, ws.Request()) {
				websocket.JSON.Send(ws, apiResponse{
					Success: false, ErrorType: errTypeUnauthorized,
					Error: errUnauthorized.Error(), Method: data.Method,
				})
				break
			}

			msg <- data
		}

		logger.Infoln("Closing web socket")
	})
}

func (con WebSocket) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}

func (con WebSocket) AuthReject(c context.Context, r *http.Request) {
	c.Set(r, readeef.CtxKey("forbidden"), true)
}

func (a apiRequest) processor(
	c context.Context,
	db readeef.DB,
	user readeef.User,
	fm *readeef.FeedManager,
	si readeef.SearchIndex,
) (Processor, error) {

	switch a.Method {
	case "get-auth-data":
		return &getAuthDataProcessor{user: user}, nil
	case "mark-article-as-read":
		return &markArticleAsReadProcessor{db: db, user: user}, nil
	case "mark-article-as-favorite":
		return &markArticleAsFavoriteProcessor{db: db, user: user}, nil
	case "format-article":
		return &formatArticleProcessor{
			db: db, user: user,
			webfwConfig:   webfw.GetConfig(c),
			readeefConfig: readeef.GetConfig(c),
		}, nil
	case "get-article":
		return &getArticleProcessor{db: db, user: user}, nil
	case "list-feeds":
		return &listFeedsProcessor{db: db, user: user}, nil
	case "discover-feeds":
		return &discoverFeedsProcessor{db: db, user: user, fm: fm}, nil
	case "parse-opml":
		return &parseOpmlProcessor{db: db, user: user, fm: fm}, nil
	case "add-feed":
		return &addFeedProcessor{db: db, user: user, fm: fm}, nil
	case "remove-feed":
		return &removeFeedProcessor{db: db, user: user, fm: fm}, nil
	case "get-feed-tags":
		return &getFeedTagsProcessor{db: db, user: user}, nil
	case "set-feed-tags":
		return &setFeedTagsProcessor{db: db, user: user}, nil
	case "mark-feed-as-read":
		return &markFeedAsReadProcessor{db: db, user: user}, nil
	case "get-feed-articles":
		return &getFeedArticlesProcessor{db: db, user: user}, nil
	case "search":
		return &searchProcessor{db: db, user: user, si: si}, nil
	case "get-user-attribute":
		return &getUserAttributeProcessor{db: db, user: user}, nil
	case "set-user-attribute":
		return &setUserAttributeProcessor{db: db, user: user}, nil
	case "list-users":
		return &listUsersProcessor{db: db, user: user}, nil
	case "add-user":
		return &addUserProcessor{db: db, user: user}, nil
	case "remove-user":
		return &removeUserProcessor{db: db, user: user}, nil
	case "set-attribute-for-user":
		return &setAttributeForUserProcessor{db: db, user: user}, nil
	default:
		return nil, errInvalidMethodValue
	}
}

func forbidden(c context.Context, r *http.Request) bool {
	if v, ok := c.Get(r, readeef.CtxKey("forbidden")); ok {
		if f, ok := v.(bool); ok {
			return f
		}
	}

	return false
}
