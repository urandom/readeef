package eventable

import (
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

const (
	ArticleStateEvent = "article-state-change"

	read  = "read"
	favor = "favor"
)

type ArticleStateData struct {
	User    content.Login          `json:"user"`
	State   string                 `json:"state"`
	Value   bool                   `json:"value"`
	Options map[string]interface{} `json:"options"`
}

func (e ArticleStateData) UserLogin() content.Login {
	return e.User
}

type articleRepo struct {
	repo.Article
	eventBus bus
	log      log.Log
}

func (r articleRepo) Read(state bool, user content.User, opts ...content.QueryOpt) error {
	err := r.Article.Read(state, user, opts...)

	if err == nil {
		r.log.Debugf("Dispatching article read state event")

		o := content.QueryOptions{}
		o.Apply(opts)

		r.eventBus.Dispatch(
			ArticleStateEvent,
			ArticleStateData{user.Login, read, state, convertOptions(o)},
		)

		r.log.Debugf("Dispatch of article read state event end")
	}

	return err
}

func (r articleRepo) Favor(state bool, user content.User, opts ...content.QueryOpt) error {
	err := r.Article.Favor(state, user, opts...)

	if err == nil {
		r.log.Debugf("Dispatching article favor state event")

		o := content.QueryOptions{}
		o.Apply(opts)

		r.eventBus.Dispatch(
			ArticleStateEvent,
			ArticleStateData{user.Login, favor, state, convertOptions(o)},
		)

		r.log.Debugf("Dispatch of article favor state event end")
	}

	return err
}

func convertOptions(o content.QueryOptions) map[string]interface{} {
	data := map[string]interface{}{}

	if o.ReadOnly {
		data["readOnly"] = true
	}

	if o.UnreadOnly {
		data["unreadOnly"] = true
	}

	if o.FavoriteOnly {
		data["favoriteOnly"] = true
	}

	if o.UntaggedOnly {
		data["untaggedOnly"] = true
	}

	if o.BeforeID > 0 {
		data["beforeID"] = o.BeforeID
	}

	if o.AfterID > 0 {
		data["afterID"] = o.AfterID
	}

	if !o.BeforeDate.IsZero() {
		data["beforeDate"] = o.BeforeDate
	}

	if !o.AfterDate.IsZero() {
		data["afterDate"] = o.AfterDate
	}

	if len(o.IDs) > 0 {
		data["ids"] = o.IDs
	}

	if len(o.FeedIDs) > 0 {
		data["feedIDs"] = o.FeedIDs
	}

	return data
}
