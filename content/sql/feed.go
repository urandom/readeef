package sql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/db"
	"github.com/urandom/webfw"
)

type Feed struct {
	base.Feed
	logger webfw.Logger

	db          *db.DB
	newArticles []content.Article
}

type UserFeed struct {
	base.UserFeed
	Feed
}

type TaggedFeed struct {
	UserFeed

	initialized bool
	tags        []content.Tag
}

type taggedFeedJSON struct {
	data.Feed
	Tags []content.Tag
}

func (f Feed) NewArticles() (a []content.Article) {
	if f.HasErr() {
		return
	}

	return f.newArticles
}

func (f *Feed) Users() (u []content.User) {
	if f.HasErr() {
		return
	}

	id := f.Data().Id
	f.logger.Infof("Getting users for feed %d\n", id)

	var in []data.User
	if err := f.db.Select(&in, f.db.SQL("get_feed_users"), id); err != nil {
		f.Err(err)
		return
	}

	u = make([]content.User, len(in))
	for i := range in {
		u[i] = f.Repo().User()
		u[i].Data(in[i])

		if u[i].HasErr() {
			f.Err(u[i].Err())
			return
		}
	}

	return
}

func (f *Feed) Update() {
	if f.HasErr() {
		return
	}

	if err := f.Validate(); err != nil {
		f.Err(err)
		return
	}

	i := f.Data()
	id := i.Id
	f.logger.Infof("Updating feed %d\n", id)

	tx, err := f.db.Begin()
	if err != nil {
		f.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(f.db.SQL("update_feed"))
	if err != nil {
		f.Err(err)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(i.Link, i.Title, i.Description, i.HubLink, i.SiteLink, i.UpdateError, i.SubscribeError, id)
	if err != nil {
		f.Err(err)
		return
	}

	if num, err := res.RowsAffected(); err != nil || num == 0 {
		id, err := f.db.CreateWithId(tx, "create_feed", i.Link, i.Title, i.Description, i.HubLink, i.SiteLink, i.UpdateError, i.SubscribeError)
		if err != nil {
			f.Err(err)
			return
		}

		i.Id = data.FeedId(id)

		pa := f.ParsedArticles()
		for index := range pa {
			ai := pa[index].Data()
			ai.FeedId = i.Id
			pa[index].Data(ai)
		}

		f.Data(i)
	}

	articles := f.updateFeedArticles(tx, f.ParsedArticles())

	if f.HasErr() {
		return
	}

	f.newArticles = articles

	tx.Commit()
}

func (f *Feed) Delete() {
	if f.HasErr() {
		return
	}

	if err := f.Validate(); err != nil {
		f.Err(err)
		return
	}

	id := f.Data().Id
	f.logger.Infof("Deleting feed %d\n", id)

	tx, err := f.db.Begin()
	if err != nil {
		f.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(f.db.SQL("delete_feed"))
	if err != nil {
		f.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		f.Err(err)
		return
	}

	tx.Commit()
}

func (f *Feed) AllArticles() (a []content.Article) {
	if f.HasErr() {
		return
	}

	id := f.Data().Id
	f.logger.Infof("Getting all feed %d articles\n", id)

	var data []data.Article
	if err := f.db.Select(&data, f.db.SQL("get_all_feed_articles"), id); err != nil {
		f.Err(err)
		return
	}

	a = make([]content.Article, len(data))
	for i := range data {
		a[i] = f.Repo().Article()
		a[i].Data(data[i])
	}

	return
}

func (f *Feed) LatestArticles() (a []content.Article) {
	if f.HasErr() {
		return
	}

	id := f.Data().Id
	f.logger.Infof("Getting latest feed %d articles\n", id)

	var data []data.Article
	if err := f.db.Select(&data, f.db.SQL("get_latest_feed_articles"), id); err != nil {
		f.Err(err)
		return
	}

	a = make([]content.Article, len(data))
	for i := range data {
		a[i] = f.Repo().Article()
		a[i].Data(data[i])
	}

	return
}

func (f *Feed) AddArticles(articles []content.Article) {
	if f.HasErr() {
		return
	}

	id := f.Data().Id
	f.logger.Infof("Adding articles to feed %d\n", id)

	tx, err := f.db.Begin()
	if err != nil {
		f.Err(err)
		return
	}
	defer tx.Rollback()

	newArticles := f.updateFeedArticles(tx, articles)

	if f.HasErr() {
		return
	}

	tx.Commit()

	f.newArticles = newArticles
}

func (f *Feed) Subscription() (s content.Subscription) {
	s = f.Repo().Subscription()
	if f.HasErr() {
		s.Err(f.Err())
		return
	}

	id := f.Data().Id
	f.logger.Infof("Getting subcription for feed %d\n", id)

	var in data.Subscription
	if err := f.db.Get(&in, f.db.SQL("get_hubbub_subscription"), id); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}
		s.Err(err)
	}

	s.Data(in)

	return
}

func (f *Feed) updateFeedArticles(tx *db.Tx, articles []content.Article) (a []content.Article) {
	if f.HasErr() || len(articles) == 0 {
		return
	}

	for i := range articles {
		if err := articles[i].Validate(); err != nil {
			f.Err(err)
			return
		}

		id := f.Data().Id
		in := articles[i].Data()
		in.FeedId = id

		var sqlString string
		args := []interface{}{in.Title, in.Description, in.Date, id}

		if in.Guid.Valid {
			sqlString = f.db.SQL("update_feed_article_with_guid")
			args = append(args, in.Guid)
		} else {
			sqlString = f.db.SQL("update_feed_article")
			args = append(args, in.Link)
		}

		stmt, err := tx.Preparex(sqlString)
		if err != nil {
			f.Err(err)
			return
		}
		defer stmt.Close()

		res, err := stmt.Exec(args...)
		if err != nil {
			f.Err(err)
			return
		}

		if num, err := res.RowsAffected(); err != nil && err == sql.ErrNoRows || num == 0 {
			a = append(a, articles[i])

			aId, err := f.db.CreateWithId(tx, "create_feed_article", id, in.Link, in.Guid,
				in.Title, in.Description, in.Date)

			if err != nil {
				f.Err(err)
				return
			}

			in.Id = data.ArticleId(aId)
			articles[i].Data(in)
		}
	}

	return
}

func (uf UserFeed) Validate() error {
	err := uf.Feed.Validate()
	if err == nil {
		err = uf.UserFeed.Validate()
	}

	return err
}

func (uf *UserFeed) Detach() {
	if uf.HasErr() {
		return
	}

	id := uf.Data().Id
	login := uf.User().Data().Login
	uf.logger.Infof("Detaching feed %d from user %s\n", id, login)

	tx, err := uf.db.Begin()
	if err != nil {
		uf.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(uf.db.SQL("delete_user_feed"))
	if err != nil {
		uf.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(login, id)
	if err != nil {
		uf.Err(err)
		return
	}

	tx.Commit()
}

func (uf *UserFeed) Articles(paging ...int) (ua []content.UserArticle) {
	if uf.HasErr() {
		return
	}

	id := uf.Data().Id
	uf.logger.Infof("Getting articles for feed %d\n", id)

	order := "read"
	articles := uf.getArticles("", order, paging...)
	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
	}

	return
}

func (uf *UserFeed) UnreadArticles(paging ...int) (ua []content.UserArticle) {
	if uf.HasErr() {
		return
	}

	id := uf.Data().Id
	uf.logger.Infof("Getting unread articles for feed %d\n", id)

	articles := uf.getArticles("ar.article_id IS NULL", "", paging...)
	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
	}

	return
}

func (uf *UserFeed) ReadBefore(date time.Time, read bool) {
	if uf.HasErr() {
		return
	}

	id := uf.Data().Id
	login := uf.User().Data().Login
	uf.logger.Infof("Marking user %s feed %d articles before %v as read: %v\n", login, id, date, read)

	tx, err := uf.db.Begin()
	if err != nil {
		uf.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(uf.db.SQL("delete_all_users_articles_read_by_feed_date"))
	if err != nil {
		uf.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(login, id, date)
	if err != nil {
		uf.Err(err)
		return
	}

	if read {
		stmt, err = tx.Preparex(uf.db.SQL("create_all_users_articles_read_by_feed_date"))
		if err != nil {
			uf.Err(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(login, id, date)
		if err != nil {
			uf.Err(err)
			return
		}
	}

	tx.Commit()
}

func (uf *UserFeed) ScoredArticles(from, to time.Time, paging ...int) (sa []content.ScoredArticle) {
	if uf.HasErr() {
		return
	}

	u := uf.User()
	id := uf.Data().Id
	login := u.Data().Login
	uf.logger.Infof("Getting scored articles for user %s feed %d between %v and %v\n", login, id, from, to)

	order := ""
	if uf.Order() == data.DescendingOrder {
		order = " DESC"
	}
	ua := getArticles(u, uf.db, uf.logger, uf, "asco.score",
		"INNER JOIN articles_scores asco ON a.id = asco.article_id",
		"uf.feed_id = $2 AND a.date > $3 AND a.date <= $4", "asco.score"+order,
		[]interface{}{id, from, to}, paging...)

	sa = make([]content.ScoredArticle, len(ua))
	for i := range ua {
		sa[i] = uf.Repo().ScoredArticle()
		sa[i].Data(ua[i].Data())
	}

	return
}

func (uf *UserFeed) getArticles(where, order string, paging ...int) (ua []content.UserArticle) {
	if uf.HasErr() {
		return
	}

	if where == "" {
		where = "uf.feed_id = $2"
	} else {
		where = "uf.feed_id = $2 AND " + where
	}

	u := uf.User()
	ua = getArticles(u, uf.db, uf.logger, uf, "", "", where, order, []interface{}{uf.Data().Id}, paging...)

	if u.HasErr() {
		uf.Err(u.Err())
	}

	return
}

func (uf *UserFeed) Query(term string, index bleve.Index, paging ...int) (ua []content.UserArticle) {
	if uf.HasErr() {
		return
	}

	var err error

	ua, err = query(term, uf.Highlight(), index, uf.User(), []data.FeedId{uf.Data().Id}, paging...)
	uf.Err(err)

	return
}

func (tf TaggedFeed) MarshalJSON() ([]byte, error) {
	tfjson := taggedFeedJSON{Feed: tf.Data(), Tags: tf.tags}
	return json.Marshal(tfjson)
}

func (tf *TaggedFeed) Tags(tags ...[]content.Tag) []content.Tag {
	if len(tags) > 0 {
		tf.tags = tags[0]
		tf.initialized = true
		return tf.tags
	}

	if !tf.initialized {
		id := tf.Data().Id
		login := tf.User().Data().Login
		tf.logger.Infof("Getting tags for user %s and feed '%d'\n", login, id)

		var feedIdTags []feedIdTag
		if err := tf.db.Select(&feedIdTags, tf.db.SQL("get_user_feed_tags"), login, id); err != nil {
			tf.Err(err)
			return []content.Tag{}
		}

		for _, t := range feedIdTags {
			tag := tf.Repo().Tag(tf.User())
			tag.Value(t.TagValue)
			tf.tags = append(tf.tags, tag)
		}
		tf.initialized = true
	}

	return tf.tags
}

func (tf *TaggedFeed) UpdateTags() {
	if tf.HasErr() {
		return
	}

	if err := tf.Validate(); err != nil {
		tf.Err(err)
		return
	}

	id := tf.Data().Id
	login := tf.User().Data().Login
	tf.logger.Infof("Deleting all tags for feed %d\n", id)

	tx, err := tf.db.Begin()
	if err != nil {
		tf.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(tf.db.SQL("delete_user_feed_tags"))
	if err != nil {
		tf.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(login, id)
	if err != nil {
		tf.Err(fmt.Errorf("Error deleting user feed tags: %v", err))
		return
	}

	tags := tf.Tags()

	if len(tags) > 0 {
		id := tf.Data().Id
		login := tf.User().Data().Login
		tf.logger.Infof("Adding tags for feed %d\n", id)

		stmt, err := tx.Preparex(tf.db.SQL("create_user_feed_tag"))
		if err != nil {
			tf.Err(err)
			return
		}
		defer stmt.Close()

		existing := tf.Tags()
		existingMap := make(map[data.TagValue]bool)

		for i := range existing {
			existingMap[existing[i].Value()] = true
		}

		for i := range tags {
			if err := tags[i].Validate(); err != nil {
				tf.Err(err)
				return
			}
			fmt.Println(tags[i])

			_, err = stmt.Exec(login, id, tags[i].Value())
			if err != nil {
				tf.Err(fmt.Errorf("Error adding user feed tag: %v", err))
				return
			}

			if !existingMap[tags[i].Value()] {
				existing = append(existing, tags[i])
			}
		}

		tf.Tags(existing)
	}

	tx.Commit()
}
