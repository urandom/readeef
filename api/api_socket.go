package api

import (
	"errors"
	"io"
	"net/http"
	"sync"

	"github.com/urandom/readeef"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"golang.org/x/net/websocket"
)

type ApiSocket struct {
	webfw.BasePatternController
	fm         *readeef.FeedManager
	si         readeef.SearchIndex
	updateFeed chan readeef.Feed
}

type apiRequest struct {
	Method    string                 `json:"method"`
	Tag       string                 `json:"tag"`
	Arguments map[string]interface{} `json:"arguments"`
}

type apiResponse struct {
	Success   bool                   `json:"success"`
	ErrorType string                 `json:"error-type"`
	Error     error                  `json:"error"`
	Method    string                 `json:"method"`
	Tag       string                 `json:"tag"`
	Arguments map[string]interface{} `json:"arguments"`
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

func NewApiSocket(fm *readeef.FeedManager, si readeef.SearchIndex) ApiSocket {
	return ApiSocket{
		BasePatternController: webfw.NewBasePatternController("/v:version/", webfw.MethodGet, ""),
		fm:         fm,
		si:         si,
		updateFeed: make(chan readeef.Feed),
	}
}

func (con ApiSocket) UpdateFeedChannel() chan<- readeef.Feed {
	return con.updateFeed
}

func (con ApiSocket) Handler(c context.Context) http.Handler {
	var mutex sync.RWMutex

	receivers := make(map[chan readeef.Feed]bool)

	go func() {
		for {
			select {
			case feed := <-con.updateFeed:
				mutex.RLock()

				readeef.Debug.Printf("Feed %s updated. Notifying %d receivers.", feed.Link, len(receivers))
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
					switch data.Method {
					case "get-auth-data":
						r = getAuthData(user)
					case "mark-article-as-read", "mark-article-as-favorite", "format-article":
						if articleId, ok := data.Arguments["id"].(float64); ok {
							switch data.Method {
							case "mark-article-as-read":
								if value, ok := data.Arguments["value"].(bool); ok {
									r = markArticleAsRead(db, user, int64(articleId), value)
								} else {
									r.err = errInvalidArgValue
									r.errType = errTypeInvalidArgValue
								}
							case "mark-article-as-favorite":
								if value, ok := data.Arguments["value"].(bool); ok {
									r = markArticleAsFavorite(db, user, int64(articleId), value)
								} else {
									r.err = errInvalidArgValue
									r.errType = errTypeInvalidArgValue
								}
							case "format-article":
								r = formatArticle(db, user, int64(articleId), webfw.GetConfig(c), readeef.GetConfig(c))
							}
						} else {
							r.err = errNoId
							r.errType = errTypeNoId
						}
					case "list-feeds":
						r = listFeeds(db, user)
					case "discover-feeds":
						if link, ok := data.Arguments["link"].(string); ok {
							r = discoverFeeds(db, user, con.fm, link)
						} else {
							r.err = errInvalidArgValue
							r.errType = errTypeInvalidArgValue
						}
					case "parse-opml":
						if opml, ok := data.Arguments["opml"].(string); ok {
							r = parseOpml(db, user, con.fm, []byte(opml))
						} else {
							r.err = errInvalidArgValue
							r.errType = errTypeInvalidArgValue
						}
					case "add-feed":
						var ok bool
						var slice []interface{}
						var links []string

						if slice, ok = data.Arguments["links"].([]interface{}); ok {
							links, ok = interfaceSliceToString(slice)
						}

						if ok {
							r = addFeed(db, user, con.fm, links)
						} else {
							r.err = errInvalidArgValue
							r.errType = errTypeInvalidArgValue
						}
					case "remove-feed":
						if id, ok := data.Arguments["id"].(float64); ok {
							r = removeFeed(db, user, con.fm, int64(id))
						} else {
							r.err = errInvalidArgValue
							r.errType = errTypeInvalidArgValue
						}
					case "get-feed-tags":
						if id, ok := data.Arguments["id"].(float64); ok {
							r = getFeedTags(db, user, int64(id))
						} else {
							r.err = errInvalidArgValue
							r.errType = errTypeInvalidArgValue
						}
					case "set-feed-tags":
						var ok bool
						var id float64
						var slice []interface{}
						var tags []string

						if id, ok = data.Arguments["id"].(float64); ok {
							if slice, ok = data.Arguments["tags"].([]interface{}); ok {
								tags, ok = interfaceSliceToString(slice)
							}
						}

						if ok {
							r = setFeedTags(db, user, int64(id), tags)
						} else {
							r.err = errInvalidArgValue
							r.errType = errTypeInvalidArgValue
						}
					case "mark-feed-as-read":
						var ok bool
						var id string
						var timestamp float64

						if id, ok = data.Arguments["id"].(string); ok {
							timestamp, ok = data.Arguments["timestamp"].(float64)
						}

						if ok {
							r = markFeedAsRead(db, user, id, int64(timestamp))
						} else {
							r.err = errInvalidArgValue
							r.errType = errTypeInvalidArgValue
						}
					case "get-feed-articles":
						var ok bool
						var id string
						var limit, offset float64
						var newerFirst, unreadOnly bool

						if id, ok = data.Arguments["id"].(string); ok {
							if limit, ok = data.Arguments["limit"].(float64); ok {
								if offset, ok = data.Arguments["offset"].(float64); ok {
									if newerFirst, ok = data.Arguments["newerFirst"].(bool); ok {
										if unreadOnly, ok = data.Arguments["unreadOnly"].(bool); ok {
										}
									}
								}
							}
						}

						if ok {
							r = getFeedArticles(db, user, id, int(limit), int(offset), newerFirst, unreadOnly)
						} else {
							r.err = errInvalidArgValue
							r.errType = errTypeInvalidArgValue
						}
					case "search":
						var ok bool
						var query, highlight, feedId string

						if query, ok = data.Arguments["query"].(string); ok {
							if highlight, ok = data.Arguments["highlight"].(string); ok {
								feedId, ok = data.Arguments["id"].(string)
							}
						}

						if ok {
							r = search(db, user, con.si, query, highlight, feedId)
						} else {
							r.err = errInvalidArgValue
							r.errType = errTypeInvalidArgValue
						}
					case "get-user-attribute":
						if attr, ok := data.Arguments["attribute"].(string); ok {
							r = getUserAttribute(db, user, attr)
						} else {
							r.err = errInvalidArgValue
							r.errType = errTypeInvalidArgValue
						}
					case "set-user-attribute":
						var ok bool
						var attr, value string

						if attr, ok = data.Arguments["attribute"].(string); ok {
							value, ok = data.Arguments["value"].(string)
						}

						if ok {
							r = setUserAttribute(db, user, attr, value)
						} else {
							r.err = errInvalidArgValue
							r.errType = errTypeInvalidArgValue
						}
					case "list-users":
						r = listUsers(db, user)
					case "add-user":
						var ok bool
						var login, password string

						if login, ok = data.Arguments["login"].(string); ok {
							password, ok = data.Arguments["password"].(string)
						}

						if ok {
							r = addUser(db, user, login, password)
						} else {
							r.err = errInvalidArgValue
							r.errType = errTypeInvalidArgValue
						}
					case "remove-user":
						if login, ok := data.Arguments["login"].(string); ok {
							r = removeUser(db, user, login)
						} else {
							r.err = errInvalidArgValue
							r.errType = errTypeInvalidArgValue
						}
					case "set-attribute-for-user":
						var ok bool
						var login, attr, value string

						if login, ok = data.Arguments["login"].(string); ok {
							if attr, ok = data.Arguments["attribute"].(string); ok {
								value, ok = data.Arguments["value"].(string)
							}
						}

						if ok {
							r = setAttributeForUser(db, user, login, attr, value)
						} else {
							r.err = errInvalidArgValue
							r.errType = errTypeInvalidArgValue
						}
					}

					go func() {
						resp <- apiResponse{
							Success: r.err == nil, Error: r.err, ErrorType: r.errType,
							Method: data.Method, Tag: data.Tag, Arguments: r.val,
						}
					}()
				case f := <-receiver:
					readeef.Debug.Println("Received notification for feed update of" + f.Link)

					r := newResponse()

					uf, _ := db.GetUserFeed(f.Id, user)

					if uf.Id > 0 {
						r.val["Feed"] = feed{
							Id: f.Id, Title: f.Title, Description: f.Description,
							Link: f.Link, Image: f.Image,
						}

						go func() {
							resp <- apiResponse{
								Success: r.err == nil, Error: r.err, ErrorType: r.errType,
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
						Error: err, Method: data.Method,
					})
				}
			}

			if forbidden(c, ws.Request()) {
				websocket.JSON.Send(ws, apiResponse{
					Success: false, ErrorType: errTypeUnauthorized,
					Error: errUnauthorized, Method: data.Method,
				})
				break
			}

			msg <- data
		}

		readeef.Debug.Println("Closing web socket")
	})
}

func (con ApiSocket) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}

func (con ApiSocket) AuthReject(c context.Context, r *http.Request) {
	c.Set(r, readeef.CtxKey("forbidden"), true)
}

func forbidden(c context.Context, r *http.Request) bool {
	if v, ok := c.Get(r, readeef.CtxKey("forbidden")); ok {
		if f, ok := v.(bool); ok {
			return f
		}
	}

	return false
}

func interfaceSliceToString(data []interface{}) (ret []string, ok bool) {
	ok = true
	ret = make([]string, len(data))

	for i := range data {
		if ret[i], ok = data[i].(string); !ok {
			return
		}
	}

	return
}
