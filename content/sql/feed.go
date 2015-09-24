package sql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/jmoiron/sqlx"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/content/sql/db"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/util"
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

type readStateInsertFeedData struct {
	NewArticleIds string
}

var (
	readStateInsertFeedTemplate *template.Template
)

func (f *Feed) ParsedArticles() (a []content.Article) {
	if f.HasErr() {
		return
	}

	articles := f.Feed.ParsedArticles()
	r := f.Repo()
	id := f.Data().Id
	a = make([]content.Article, len(articles))

	for i := range articles {
		article := r.Article()
		data := articles[i].Data()
		data.FeedId = id
		article.Data(data)
		a[i] = article
	}

	return
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
	if id == 0 {
		f.Err(content.NewValidationError(errors.New("Invalid feed id")))
		return
	}

	f.logger.Infof("Getting users for feed %d\n", id)

	var in []data.User
	if err := f.db.Select(&in, f.db.SQL().Feed.GetUsers, id); err != nil {
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
	s := f.db.SQL()
	f.logger.Infof("Updating feed %d\n", id)

	tx, err := f.db.Beginx()
	if err != nil {
		f.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(s.Feed.Update)
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
		id, err := f.db.CreateWithId(tx, s.Feed.Create, i.Link, i.Title, i.Description, i.HubLink, i.SiteLink, i.UpdateError, i.SubscribeError)
		if err != nil {
			f.Err(err)
			return
		}

		i.Id = data.FeedId(id)

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
	if id == 0 {
		f.Err(content.NewValidationError(errors.New("Invalid feed id")))
		return
	}

	f.logger.Infof("Deleting feed %d\n", id)

	tx, err := f.db.Beginx()
	if err != nil {
		f.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(f.db.SQL().Feed.Delete)
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

func (f *Feed) SetNewArticlesUnread() {
	if f.HasErr() {
		return
	}

	if err := f.Validate(); err != nil {
		f.Err(err)
		return
	}

	id := f.Data().Id
	if id == 0 {
		f.Err(content.NewValidationError(errors.New("Invalid feed id")))
		return
	}

	articles := f.NewArticles()
	if len(articles) == 0 {
		return
	}

	var err error
	if readStateInsertFeedTemplate == nil {
		readStateInsertFeedTemplate, err = template.New("read-state-insert-feed-sql").
			Parse(f.db.SQL().Feed.ReadStateInsertTemplate)

		if err != nil {
			f.Err(fmt.Errorf("Error generating read-state-insert-feed template: %v", err))
			return
		}
	}

	f.logger.Infof("Setting new articles from feed %d as unread for all related users\n", id)

	args := make([]interface{}, len(articles)+1)
	placeholders := make([]string, len(articles))

	args[0] = id
	for i := range articles {
		args[i+1] = articles[i].Data().Id
		placeholders[i] = fmt.Sprintf("$%d", i+2)
	}

	buf := util.BufferPool.GetBuffer()
	defer util.BufferPool.Put(buf)

	data := readStateInsertFeedData{NewArticleIds: strings.Join(placeholders, ", ")}

	if err := readStateInsertFeedTemplate.Execute(buf, data); err != nil {
		f.Err(fmt.Errorf("Error executing read-state-insert-feed template: %v", err))
		return
	}

	sql := buf.String()
	f.logger.Debugf("Read state insert feed SQL:\n%s\nArgs:%q\n", sql, args)

	tx, err := f.db.Beginx()
	if err != nil {
		f.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(sql)
	if err != nil {
		f.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		f.Err(err)
		return
	}

	if err := tx.Commit(); err != nil {
		f.Err(err)
	}
}

func (f *Feed) AllArticles() (a []content.Article) {
	if f.HasErr() {
		return
	}

	id := f.Data().Id
	if id == 0 {
		f.Err(content.NewValidationError(errors.New("Invalid feed id")))
		return
	}

	f.logger.Infof("Getting all feed %d articles\n", id)

	var data []data.Article
	if err := f.db.Select(&data, f.db.SQL().Feed.GetAllArticles, id); err != nil {
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
	if id == 0 {
		f.Err(content.NewValidationError(errors.New("Invalid feed id")))
		return
	}

	f.logger.Infof("Getting latest feed %d articles\n", id)

	var data []data.Article
	if err := f.db.Select(&data, f.db.SQL().Feed.GetLatestArticles, id); err != nil {
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
	if id == 0 {
		f.Err(content.NewValidationError(errors.New("Invalid feed id")))
		return
	}

	f.logger.Infof("Adding %d articles to feed %d\n", len(articles), id)

	tx, err := f.db.Beginx()
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
	if id == 0 {
		f.Err(content.NewValidationError(errors.New("Invalid feed id")))
		return
	}

	f.logger.Infof("Getting subcription for feed %d\n", id)

	var in data.Subscription
	if err := f.db.Get(&in, f.db.SQL().Feed.GetHubbubSubscription, id); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}
		s.Err(err)
	}

	in.FeedId = id
	s.Data(in)

	return
}

func (f *Feed) updateFeedArticles(tx *sqlx.Tx, articles []content.Article) (a []content.Article) {
	if f.HasErr() || len(articles) == 0 {
		return
	}

	id := f.Data().Id

	for i := range articles {
		d := articles[i].Data()
		d.FeedId = id
		articles[i].Data(d)

		updateArticle(articles[i], tx, f.db, f.logger)

		if articles[i].HasErr() {
			f.Err(fmt.Errorf("Error updating article %s: %v\n", articles[i], articles[i].Err()))
			return
		}

		if articles[i].Data().IsNew {
			a = append(a, articles[i])
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

	if err := uf.Validate(); err != nil {
		uf.Err(err)
		return
	}

	id := uf.Data().Id
	if id == 0 {
		uf.Err(content.NewValidationError(errors.New("Invalid feed id")))
		return
	}

	login := uf.User().Data().Login
	uf.logger.Infof("Detaching feed %d from user %s\n", id, login)

	tx, err := uf.db.Beginx()
	if err != nil {
		uf.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(uf.db.SQL().Feed.Detach)
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

func (uf *UserFeed) Articles(o ...data.ArticleQueryOptions) (ua []content.UserArticle) {
	if uf.HasErr() {
		return
	}

	if err := uf.Validate(); err != nil {
		uf.Err(err)
		return
	}

	id := uf.Data().Id
	if id == 0 {
		uf.Err(content.NewValidationError(errors.New("Invalid feed id")))
		return
	}

	var opts data.ArticleQueryOptions
	if len(o) > 0 {
		opts = o[0]
	}

	uf.logger.Infof("Getting articles for feed %d with options: %#v\n", id, opts)

	where := "uf.feed_id = $2"

	u := uf.User()
	ua = getArticles(u, uf.db, uf.logger, opts, uf, "", where, []interface{}{uf.Data().Id})

	if u.HasErr() {
		uf.Err(u.Err())
	}

	return
}

func (uf *UserFeed) Ids(o ...data.ArticleIdQueryOptions) (ids []data.ArticleId) {
	if uf.HasErr() {
		return
	}

	if err := uf.Validate(); err != nil {
		uf.Err(err)
		return
	}

	u := uf.User()
	id := uf.Data().Id

	var opts data.ArticleIdQueryOptions
	if len(o) > 0 {
		opts = o[0]
	}

	uf.logger.Infof("Getting user %s feed %d article ids with options %#v\n", u.Data().Login, id, opts)

	ids = articleIds(u, uf.db, uf.logger, opts, "", "a.feed_id = $2", []interface{}{id})

	if u.HasErr() {
		uf.Err(u.Err())
	}

	return
}

func (uf *UserFeed) Count(o ...data.ArticleCountOptions) (count int64) {
	if uf.HasErr() {
		return
	}

	if err := uf.Validate(); err != nil {
		uf.Err(err)
		return
	}

	u := uf.User()
	id := uf.Data().Id

	var opts data.ArticleCountOptions
	if len(o) > 0 {
		opts = o[0]
	}

	uf.logger.Infof("Getting user %s feed %d article count with options %#v\n", u.Data().Login, id, opts)

	count = articleCount(u, uf.db, uf.logger, opts, "", "a.feed_id = $2", []interface{}{id})

	if u.HasErr() {
		uf.Err(u.Err())
	}

	return
}

func (uf *UserFeed) ReadState(read bool, o ...data.ArticleUpdateStateOptions) {
	if uf.HasErr() {
		return
	}

	if err := uf.Validate(); err != nil {
		uf.Err(err)
		return
	}

	var opts data.ArticleUpdateStateOptions
	if len(o) > 0 {
		opts = o[0]
	}

	u := uf.User()
	login := u.Data().Login
	id := uf.Data().Id
	uf.logger.Infof("Getting articles for user %s feed %d with opts: %#v\n", login, id, opts)

	args := []interface{}{id}
	readState(u, uf.db, uf.logger, opts, read, "", "uf.feed_id = $2",
		"", "a.feed_id = $2", args, args)
	if u.HasErr() {
		uf.Err(u.Err())
	}
}

func (uf *UserFeed) Query(term string, sp content.SearchProvider, paging ...int) (ua []content.UserArticle) {
	if uf.HasErr() {
		return
	}

	if err := uf.Validate(); err != nil {
		uf.Err(err)
		return
	}

	id := uf.Data().Id
	if id == 0 {
		uf.Err(content.NewValidationError(errors.New("Invalid feed id")))
		return
	}

	var err error

	limit, offset := pagingLimit(paging)
	ua, err = sp.Search(term, uf.User(), []data.FeedId{id}, limit, offset)
	uf.Err(err)

	return
}

func (tf TaggedFeed) MarshalJSON() ([]byte, error) {
	tfjson := taggedFeedJSON{Feed: tf.Data(), Tags: tf.tags}
	b, err := json.Marshal(tfjson)

	if err == nil {
		return b, nil
	} else {
		return []byte{}, fmt.Errorf("Error marshaling tagged feed data for %s: %v", tf, err)
	}
}

func (tf *TaggedFeed) Tags(tags ...[]content.Tag) []content.Tag {
	if len(tags) > 0 {
		tf.tags = tags[0]
		tf.initialized = true
		return tf.tags
	}

	if !tf.initialized {
		if err := tf.Validate(); err != nil {
			tf.Err(err)
			return tf.tags
		}

		id := tf.Data().Id
		if id == 0 {
			tf.Err(content.NewValidationError(errors.New("Invalid feed id")))
			return tf.tags
		}

		login := tf.User().Data().Login
		tf.logger.Infof("Getting tags for user %s and feed '%d'\n", login, id)

		var tagData []data.Tag
		if err := tf.db.Select(&tagData, tf.db.SQL().Feed.GetUserTags, login, id); err != nil {
			tf.Err(err)
			return []content.Tag{}
		}

		for _, d := range tagData {
			tag := tf.Repo().Tag(tf.User())
			tag.Data(d)
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
	s := tf.db.SQL()
	u := tf.User()
	if id == 0 {
		tf.Err(content.NewValidationError(errors.New("Invalid feed id")))
		return
	}

	login := u.Data().Login
	tf.logger.Infof("Deleting all tags for feed %d\n", id)

	tx, err := tf.db.Beginx()
	if err != nil {
		tf.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(s.Feed.DeleteUserTags)
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
	for i := range tags {
		tagData := tags[i].Data()
		if tagData.Id == 0 {
			t := u.TagByValue(tagData.Value)
			if t.HasErr() {
				err := t.Err()
				if err != content.ErrNoContent {
					tf.Err(t.Err())
					return
				}
			} else {
				tagData.Id = t.Data().Id
				tags[i].Data(tagData)
			}
		}
	}

	if len(tags) > 0 {
		id := tf.Data().Id
		tf.logger.Infof("Adding tags for feed %d\n", id)

		stmt, err := tx.Preparex(s.Feed.CreateUserTag)
		if err != nil {
			tf.Err(err)
			return
		}
		defer stmt.Close()

		for i := range tags {
			if err := tags[i].Validate(); err != nil {
				tf.Err(err)
				return
			}

			tagData := tags[i].Data()

			if tagData.Id == 0 {
				tagId, err := tf.db.CreateWithId(tx, s.Tag.Create, tagData.Value)
				if err != nil {
					tf.Err(err)
					return
				}
				tagData.Id = data.TagId(tagId)
			}

			_, err = stmt.Exec(login, id, tagData.Id)
			if err != nil {
				tf.Err(fmt.Errorf("Error adding user feed tag for user %s, feed %d, and tag %s: %v", login, id, tagData.Value, err))
				return
			}
		}
	}

	stmt, err = tx.Preparex(s.Tag.DeleteStale)
	if err != nil {
		tf.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		tf.Err(fmt.Errorf("Error deleting stale tags: %v", err))
		return
	}

	tx.Commit()
}
