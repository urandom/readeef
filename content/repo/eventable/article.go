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
	User    content.Login        `json:"user"`
	State   string               `json:"state"`
	Value   bool                 `json:"value"`
	Options content.QueryOptions `json:"options"`
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
	o := content.QueryOptions{}
	o.Apply(opts)

	err := r.Article.Read(state, user, opts...)

	if err == nil {
		r.log.Debugf("Logging article read state event")
		r.eventBus.Dispatch(
			ArticleStateEvent,
			ArticleStateData{user.Login, read, state, o},
		)
	}

	return err
}

func (r articleRepo) Favor(state bool, user content.User, opts ...content.QueryOpt) error {
	o := content.QueryOptions{}
	o.Apply(opts)

	err := r.Article.Favor(state, user, opts...)

	if err == nil {
		r.log.Debugf("Logging article favor state event")
		r.eventBus.Dispatch(
			ArticleStateEvent,
			ArticleStateData{user.Login, read, state, o},
		)
	}

	return err
}
