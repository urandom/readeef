package sql

import (
	"database/sql"
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/content/sql/db"
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

func (r *Repo) UserByLogin(login data.Login) (u content.User) {
	u = r.User()
	if r.HasErr() {
		u.Err(r.Err())
		return
	}

	r.logger.Infof("Getting user '%s'\n", login)

	var data data.User
	if err := r.db.Get(&data, r.db.SQL().Repo.GetUser, login); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}
		u.Err(err)
		return
	}

	data.Login = login
	u.Data(data)

	return
}

func (r *Repo) UserByMD5Api(md5 []byte) (u content.User) {
	u = r.User()
	if r.HasErr() {
		u.Err(r.Err())
		return
	}

	r.logger.Infof("Getting user using md5 api field '%v'\n", md5)

	var d data.User
	if err := r.db.Get(&d, r.db.SQL().Repo.GetUserByMD5API, md5); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}
		u.Err(err)
		return
	}

	d.MD5API = md5
	u.Data(d)

	return
}

func (r *Repo) AllUsers() (users []content.User) {
	if r.HasErr() {
		return
	}

	r.logger.Infoln("Getting all users")

	var data []data.User
	if err := r.db.Select(&data, r.db.SQL().Repo.GetUsers); err != nil {
		r.Err(err)
		return
	}

	users = make([]content.User, len(data))

	for i := range data {
		users[i] = r.User()
		users[i].Data(data[i])
		if users[i].HasErr() {
			r.Err(users[i].Err())
			return
		}
	}

	return
}

func (r *Repo) FeedById(id data.FeedId) (f content.Feed) {
	f = r.Feed()
	if r.HasErr() {
		f.Err(r.Err())
		return
	}

	r.logger.Infof("Getting feed '%d'\n", id)

	i := data.Feed{}
	if err := r.db.Get(&i, r.db.SQL().Repo.GetFeed, id); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}
		f.Err(err)
		return
	}

	i.Id = id
	f.Data(i)

	return
}

func (r *Repo) FeedByLink(link string) (f content.Feed) {
	f = r.Feed()
	if r.HasErr() {
		f.Err(r.Err())
		return
	}

	r.logger.Infof("Getting feed by link '%s'\n", link)

	i := data.Feed{}
	if err := r.db.Get(&i, r.db.SQL().Repo.GetFeedByLink, link); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}
		f.Err(err)
		return
	}

	i.Link = link
	f.Data(i)

	return
}

func (r *Repo) AllFeeds() (feeds []content.Feed) {
	if r.HasErr() {
		return
	}

	r.logger.Infoln("Getting all feeds")

	var data []data.Feed
	if err := r.db.Select(&data, r.db.SQL().Repo.GetFeeds); err != nil {
		r.Err(err)
		return
	}

	feeds = make([]content.Feed, len(data))

	for i := range data {
		feeds[i] = r.Feed()
		feeds[i].Data(data[i])
	}

	return
}

func (r *Repo) AllUnsubscribedFeeds() (feeds []content.Feed) {
	if r.HasErr() {
		return
	}

	r.logger.Infoln("Getting all unsubscribed feeds")

	var data []data.Feed
	if err := r.db.Select(&data, r.db.SQL().Repo.GetUnsubscribedFeeds); err != nil {
		r.Err(err)
		return
	}

	feeds = make([]content.Feed, len(data))

	for i := range data {
		feeds[i] = r.Feed()
		feeds[i].Data(data[i])
	}

	return
}

func (r *Repo) AllSubscriptions() (s []content.Subscription) {
	if r.HasErr() {
		return
	}

	r.logger.Infoln("Getting all subscriptions")

	var data []data.Subscription
	if err := r.db.Select(&data, r.db.SQL().Repo.GetHubbubSubscriptions); err != nil {
		r.Err(err)
		return
	}

	s = make([]content.Subscription, len(data))

	for i := range data {
		s[i] = r.Subscription()
		s[i].Data(data[i])
	}

	return
}

func (r *Repo) FailSubscriptions() {
	if r.HasErr() {
		return
	}

	r.logger.Infoln("Marking all subscriptions as failed")

	if _, err := r.db.Exec(r.db.SQL().Repo.FailHubbubSubscriptions); err != nil {
		r.Err(err)
		return
	}
}

func (r *Repo) DeleteStaleUnreadRecords() {
	if r.HasErr() {
		return
	}

	r.logger.Infoln("Deleting stale article unread records")

	if _, err := r.db.Exec(r.db.SQL().Repo.DeleteStaleUnreadRecords, time.Now().AddDate(0, -1, 0)); err != nil {
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
