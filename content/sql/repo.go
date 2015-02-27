package sql

import (
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/info"
	"github.com/urandom/readeef/db"
	"github.com/urandom/webfw"
)

type Repo struct {
	base.Repo
	logger webfw.Logger

	db *db.DB
}

func NewRepo(db *db.DB, logger webfw.Logger) *Repo {
	return &Repo{db: db, logger: logger}
}

func (r *Repo) UserByLogin(login info.Login) (u content.User) {
	u = r.User()
	if r.HasErr() {
		return
	}

	r.logger.Infof("Getting user '%s'\n", login)

	var info info.User
	if err := r.db.Get(&info, r.db.SQL("get_user"), login); err != nil {
		r.Err(err)
		return
	}

	info.Login = login
	u.Info(info)

	if u.HasErr() {
		r.Err(u.Err())
	}

	return
}

func (r *Repo) UserByMD5Api(md5 []byte) (u content.User) {
	u = r.User()
	if r.HasErr() {
		return
	}

	r.logger.Infof("Getting user using md5 api field '%v'\n", md5)

	var info info.User
	if err := r.db.Get(&info, r.db.SQL("get_user_by_md5_api"), md5); err != nil {
		r.Err(err)
		return
	}

	info.MD5API = md5
	u.Info(info)

	if u.HasErr() {
		r.Err(u.Err())
	}

	return
}

func (r *Repo) AllUsers() (users []content.User) {
	if r.HasErr() {
		return
	}

	r.logger.Infoln("Getting all users")

	var info []info.User
	if err := r.db.Select(&info, r.db.SQL("get_users")); err != nil {
		r.Err(err)
		return
	}

	users = make([]content.User, len(info))

	for i := range info {
		users[i] = r.User()
		users[i].Info(info[i])
		if users[i].HasErr() {
			r.Err(users[i].Err())
			return
		}
	}

	return
}

func (r *Repo) FeedById(id info.FeedId) (f content.Feed) {
	f = r.Feed()
	if r.HasErr() {
		return
	}

	r.logger.Infof("Getting feed '%d'\n", id)

	i := info.Feed{}
	if err := r.db.Get(&i, r.db.SQL("get_feed"), id); err != nil {
		r.Err(err)
		return
	}

	i.Id = id
	f.Info(i)

	return
}

func (r *Repo) FeedByLink(link string) (f content.Feed) {
	f = r.Feed()
	if r.HasErr() {
		return
	}

	r.logger.Infof("Getting feed by link '%s'\n", link)

	i := info.Feed{}
	if err := r.db.Get(&i, r.db.SQL("get_feed_by_link"), link); err != nil {
		r.Err(err)
		return
	}

	i.Link = link
	f.Info(i)

	return
}

func (r *Repo) AllFeeds() (feeds []content.Feed) {
	if r.HasErr() {
		return
	}

	r.logger.Infoln("Getting all feeds")

	var info []info.Feed
	if err := r.db.Select(&info, r.db.SQL("get_feeds")); err != nil {
		r.Err(err)
		return
	}

	feeds = make([]content.Feed, len(info))

	for i := range info {
		feeds[i] = r.Feed()
		feeds[i].Info(info[i])
	}

	return
}

func (r *Repo) AllUnsubscribedFeeds() (feeds []content.Feed) {
	if r.HasErr() {
		return
	}

	r.logger.Infoln("Getting all unsubscribed feeds")

	var info []info.Feed
	if err := r.db.Select(&info, r.db.SQL("get_unsubscribed_feeds")); err != nil {
		r.Err(err)
		return
	}

	feeds = make([]content.Feed, len(info))

	for i := range info {
		feeds[i] = r.Feed()
		feeds[i].Info(info[i])
	}

	return
}

func (r *Repo) AllSubscriptions() (s []content.Subscription) {
	if r.HasErr() {
		return
	}

	r.logger.Infoln("Getting all subscriptions")

	var info []info.Subscription
	if err := r.db.Select(&info, r.db.SQL("get_hubbub_subscriptions")); err != nil {
		r.Err(err)
		return
	}

	s = make([]content.Subscription, len(info))

	for i := range info {
		s[i] = r.Subscription()
		s[i].Info(info[i])
	}

	return
}

func (r *Repo) FailSubscriptions() {
	if r.HasErr() {
		return
	}

	r.logger.Infoln("Marking all subscriptions as failed")

	if _, err := r.db.Exec(r.db.SQL("fail_hubbub_subscriptions")); err != nil {
		r.Err(err)
		return
	}
}

func pagingLimit(paging []int) (int, int) {
	limit := 50
	offset := 0

	if len(paging) > 0 {
		limit = paging[0]
		if len(paging) > 1 {
			offset = paging[1]
		}
	}

	return limit, offset
}
