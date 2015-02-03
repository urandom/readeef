package api

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/urandom/readeef"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"golang.org/x/net/websocket"
)

type ApiSocket struct {
	webfw.BasePatternController
	fm *readeef.FeedManager
	si readeef.SearchIndex
}

type apiRequest struct {
	Method    string                 `json:"method"`
	Arguments map[string]interface{} `json:"arguments"`
}

type apiResponse struct {
	Success   bool                   `json:"success"`
	ErrorType string                 `json:"error-type"`
	Error     error                  `json:"error"`
	Method    string                 `json:"method"`
	Arguments map[string]interface{} `json:"arguments"`
}

var (
	errTypeInternal           = "error-internal"
	errTypeMessageParse       = "error-message-parse"
	errTypeNoId               = "error-no-id"
	errTypeInvalidMethodValue = "error-invalid-method-value"
	errTypeInvalidArgValue    = "error-invalid-arg-value"

	errNoId               = errors.New("No Id given")
	errInvalidMethodValue = errors.New("Invalid method")
	errInvalidArgValue    = errors.New("Invalid argument value")
	errInternal           = errors.New("Internal server error")
)

func NewApiSocket(fm *readeef.FeedManager, si readeef.SearchIndex) ApiSocket {
	return ApiSocket{
		BasePatternController: webfw.NewBasePatternController("/v:version/", webfw.MethodGet, ""),
		fm: fm,
		si: si,
	}
}

func (con ApiSocket) Handler(c context.Context) http.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		db := readeef.GetDB(c)
		user := readeef.GetUser(c, ws.Request())

		msg := make(chan apiRequest)
		resp := make(chan apiResponse)

		done := make(chan bool)
		defer close(done)

		go func() {
			for {
				var r responseError

				select {
				case data := <-msg:
					switch data.Method {
					case "get-auth-data":
						r = getAuthData(user)
						resp <- apiResponse{
							Success: r.err == nil, Error: r.err, Method: data.Method, Arguments: r.val,
						}
					case "mark-article-as-read", "mark-article-as-favorite", "format-article":
						if articleId, ok := data.Arguments["id"].(float64); ok {
							var r responseError
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
						if data, ok := data.Arguments["opml"].(string); ok {
							r = parseOpml(db, user, con.fm, strings.NewReader(data))
						} else {
							r.err = errInvalidArgValue
							r.errType = errTypeInvalidArgValue
						}
					case "add-feed":
						if links, ok := data.Arguments["links"].([]string); ok {
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
						var tags []string

						if id, ok = data.Arguments["id"].(float64); ok {
							tags, ok = data.Arguments["tags"].([]string)
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
							timestamp, ok = data.Arguments["float64"].(float64)
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
							if highlight, ok = data.Arguments["query"].(string); ok {
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
							r = setUserAttribute(db, user, attr, strings.NewReader(value))
						} else {
							r.err = errInvalidArgValue
							r.errType = errTypeInvalidArgValue
						}
					case "list-users":
						r = listUsers(db, user)
					case "add-user":
						var ok bool
						var login, userData string

						if login, ok = data.Arguments["login"].(string); ok {
							userData, ok = data.Arguments["userData"].(string)
						}

						if ok {
							r = addUser(db, user, login, strings.NewReader(userData))
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
							r = setUserAdminAttribute(db, user, login, attr, value)
						} else {
							r.err = errInvalidArgValue
							r.errType = errTypeInvalidArgValue
						}
					}

					go func() {
						resp <- apiResponse{
							Success: r.err == nil, Error: r.err, ErrorType: r.errType, Method: data.Method, Arguments: r.val,
						}
					}()
				case <-done:
					return
				}
			}
		}()

		for {
			var data apiRequest
			if err := websocket.JSON.Receive(ws, &data); err != nil {
				if err == io.EOF {
					break
				} else {
					websocket.JSON.Send(ws, apiResponse{
						Success: false, ErrorType: errTypeMessageParse,
						Error: err, Method: data.Method,
					})
				}
			}

			msg <- data
		}

		readeef.Debug.Println("Closing web socket")
	})
}
