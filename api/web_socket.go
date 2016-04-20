package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"golang.org/x/net/websocket"
)

type WebSocket struct {
	webfw.BasePatternController
	fm           *readeef.FeedManager
	sp           content.SearchProvider
	extractor    content.Extractor
	capabilities capabilities
	updateFeed   chan content.Feed
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

type heartbeatProcessor struct{}

var (
	errTypeInternal           = "error-internal"
	errTypeMessageParse       = "error-message-parse"
	errTypeNoId               = "error-no-id"
	errTypeInvalidMethodValue = "error-invalid-method-value"
	errTypeInvalidArgValue    = "error-invalid-arg-value"
	errTypeUnauthorized       = "error-unauthorized"
	errTypeResourceNotFound   = "error-resource-not-found"

	errNoId               = errors.New("No Id given")
	errInvalidMethodValue = errors.New("Invalid method")
	errInvalidArgValue    = errors.New("Invalid argument value")
	errInternal           = errors.New("Internal server error")
	errUnauthorized       = errors.New("Unauthorized")
	errResourceNotFound   = errors.New("Resource not found")
)

func NewWebSocket(fm *readeef.FeedManager, sp content.SearchProvider,
	extractor content.Extractor, capabilities capabilities) WebSocket {
	return WebSocket{
		BasePatternController: webfw.NewBasePatternController("/v:version/", webfw.MethodGet, ""),
		fm:           fm,
		sp:           sp,
		extractor:    extractor,
		capabilities: capabilities,
		updateFeed:   make(chan content.Feed),
	}
}

func (con WebSocket) FeedUpdated(f content.Feed) error {
	con.updateFeed <- f
	return nil
}

func (con WebSocket) FeedDeleted(f content.Feed) error {
	return nil
}

func (con WebSocket) Handler(c context.Context) http.Handler {
	var mutex sync.RWMutex

	receivers := make(map[chan content.Feed]bool)
	logger := webfw.GetLogger(c)

	go func() {
		for {
			select {
			case feed := <-con.updateFeed:
				logger.Infoln("New articles notification for " + feed.String())
				mutex.RLock()

				for receiver, _ := range receivers {
					receiver <- feed
				}

				mutex.RUnlock()
			}
		}
	}()

	cfg := readeef.GetConfig(c)
	return websocket.Handler(func(ws *websocket.Conn) {
		user := readeef.GetUser(c, ws.Request())
		sess := webfw.GetSession(c, ws.Request())

		msg := make(chan apiRequest)
		resp := make(chan apiResponse)

		receiver := make(chan content.Feed)

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

					if processor, err = data.processor(c, sess, user, con.fm, con.sp, con.extractor,
						con.capabilities, []byte(cfg.Auth.Secret)); err == nil {
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
							} else if err == content.ErrNoContent {
								r.err = errResourceNotFound
								r.errType = errTypeResourceNotFound
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
					if f == nil || user == nil {
						// Socket was closed
						return
					}
					logger.Infoln("Received notification for feed update of " + f.String())

					r := newResponse()

					uf := user.FeedById(f.Data().Id)

					if !uf.HasErr() {
						r.val["Feed"] = uf

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
	s context.Session,
	user content.User,
	fm *readeef.FeedManager,
	sp content.SearchProvider,
	extractor content.Extractor,
	capabilities capabilities,
	secret []byte,
) (Processor, error) {

	switch a.Method {
	case "heartbeat":
		return &heartbeatProcessor{}, nil
	case "get-auth-data":
		return &getAuthDataProcessor{user: user, session: s, capabilities: capabilities}, nil
	case "logout":
		return &logoutProcessor{session: s}, nil
	case "article-read-state":
		return &articleReadStateProcessor{user: user}, nil
	case "article-favorite-state":
		return &articleFavoriteStateProcessor{user: user}, nil
	case "format-article":
		return &formatArticleProcessor{
			user:          user,
			extractor:     extractor,
			webfwConfig:   webfw.GetConfig(c),
			readeefConfig: readeef.GetConfig(c),
		}, nil
	case "get-article":
		return &getArticleProcessor{user: user}, nil
	case "list-feeds":
		return &listFeedsProcessor{user: user}, nil
	case "discover-feeds":
		return &discoverFeedsProcessor{user: user, fm: fm}, nil
	case "export-opml":
		return &exportOpmlProcessor{user: user}, nil
	case "parse-opml":
		return &parseOpmlProcessor{user: user, fm: fm}, nil
	case "add-feeds":
		return &addFeedsProcessor{user: user, fm: fm}, nil
	case "remove-feed":
		return &removeFeedProcessor{user: user, fm: fm}, nil
	case "get-feed-tags":
		return &getFeedTagsProcessor{user: user}, nil
	case "set-feed-tags":
		return &setFeedTagsProcessor{user: user}, nil
	case "read-state":
		return &readStateProcessor{user: user}, nil
	case "get-feed-articles":
		return &getFeedArticlesProcessor{user: user, sp: sp}, nil
	case "get-user-attribute":
		return &getUserAttributeProcessor{user: user}, nil
	case "set-user-attribute":
		return &setUserAttributeProcessor{user: user, secret: secret}, nil
	case "list-users":
		return &listUsersProcessor{user: user}, nil
	case "add-user":
		return &addUserProcessor{user: user, secret: secret}, nil
	case "remove-user":
		return &removeUserProcessor{user: user}, nil
	case "set-attribute-for-user":
		return &setAttributeForUserProcessor{user: user, secret: secret}, nil
	default:
		return nil, errInvalidMethodValue
	}
}

func (p heartbeatProcessor) Process() responseError {
	return responseError{}
}

func forbidden(c context.Context, r *http.Request) bool {
	if v, ok := c.Get(r, readeef.CtxKey("forbidden")); ok {
		if f, ok := v.(bool); ok {
			return f
		}
	}

	return false
}
