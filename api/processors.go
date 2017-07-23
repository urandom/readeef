package api

/*
import "fmt"

type relativeLinkErr string

func (e relativeLinkErr) Error() string {
	return fmt.Sprintf("relative link: %s", e)
}

func (e relativeLinkErr) RelativeLink() string {
	return string(e)
}

func RelativeLinkError(err error) (string, bool) {
	type checker interface {
		RelativeLink() string
	}
	if m, ok := err.(checker); ok {
		return m.RelativeLink(), ok
	}

	return "", false
}

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
)

type Processor interface {
	Process() (args, error)
}

type ProcessorMux struct {
	user           content.User
	secret         []byte
	capabilities   capabilities
	extractor      content.Extractor
	searchProvider content.SearchProvider
	tokenStorage   content.TokenStorage
	fm             *readeef.FeedManager
}

func NewProcessorMux(
	user content.User,
	secret []byte,
	capabilities capabilities,
	extractor content.Extractor,
	searchProvider content.SearchProvider,
	tokenStorage content.TokenStorage,
	fm *readeef.FeedManager,
) ProcessorMux {
	return ProcessorMux{
		user:           user,
		secret:         secret,
		capabilities:   capabilities,
		extractor:      extractor,
		searchProvider: searchProvider,
		tokenStorage:   tokenStorage,
		fm:             fm,
	}
}

func (m ProcessorMux) Serve(method string, data []byte) (args args, terminate bool, err error) {
	processor := m.processor(method)
	if processor == nil {
		err = errors.WithStack(unknownMethodErr(method))
		return nil, false, err
	}

	if len(data) > 0 {
		if err = json.Unmarshal(data, processor); err != nil {
			err = errors.Wrapf(err, "unmarshaling data '%s' from method %s", data, method)
			return nil, false, err
		}
	}

	args, err = processor.Process()
	if err == nil {
		_, terminate = processor.(logoutProcessor)
	}

	return args, terminate, err
}

func (m ProcessorMux) processor(method string) Processor {
	switch method {
	case "logout":
		return logoutProcessor{storage: m.tokenStorage}
	case "heartbeat":
		return heartbeatProcessor{}
	case "get-auth-data":
		return getAuthDataProcessor{user: m.user, capabilities: m.capabilities}
	case "article-read-state":
		return articleReadStateProcessor{user: m.user}
	case "article-favorite-state":
		return articleFavoriteStateProcessor{user: m.user}
	case "format-article":
		return formatArticleProcessor{user: m.user, extractor: m.extractor}
	case "get-article":
		return getArticleProcessor{user: m.user}
	case "list-feeds":
		return listFeedsProcessor{user: m.user}
	case "discover-feeds":
		return discoverFeedsProcessor{user: m.user, fm: m.fm}
	case "export-opml":
		return exportOpmlProcessor{user: m.user}
	case "parse-opml":
		return parseOpmlProcessor{user: m.user, fm: m.fm}
	case "add-feeds":
		return addFeedsProcessor{user: m.user, fm: m.fm}
	case "remove-feed":
		return removeFeedProcessor{user: m.user, fm: m.fm}
	case "get-feed-tags":
		return getFeedTagsProcessor{user: m.user}
	case "set-feed-tags":
		return setFeedTagsProcessor{user: m.user}
	case "read-state":
		return readStateProcessor{user: m.user}
	case "get-feed-articles":
		return getFeedArticlesProcessor{user: m.user, sp: m.searchProvider}
	case "get-user-attribute":
		return getUserAttributeProcessor{user: m.user}
	case "set-user-attribute":
		return setUserAttributeProcessor{user: m.user, secret: m.secret}
	case "list-users":
		return listUsersProcessor{user: m.user}
	case "add-user":
		return addUserProcessor{user: m.user, secret: m.secret}
	case "remove-user":
		return removeUserProcessor{user: m.user}
	case "set-attribute-for-user":
		return setAttributeForUserProcessor{user: m.user, secret: m.secret}
	default:
		return nil
	}
}

type unknownMethodErr string

func (e unknownMethodErr) Error() string {
	return fmt.Sprintf("unknown method: %s", e)
}

func (e unknownMethodErr) InvalidMethod() string {
	return string(e)
}

func InvalidMethodError(err error) (string, bool) {
	type checker interface {
		InvalidMethod() string
	}

	if m, ok := err.(checker); ok {
		return m.InvalidMethod(), ok
	}

	return "", false
}

type heartbeatProcessor struct{}

func (p heartbeatProcessor) Process() (args, error) {
	return nil, nil
}
*/
