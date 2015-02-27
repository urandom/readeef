package sql

import (
	dsql "database/sql"
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/info"
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
	base.TaggedFeed
	UserFeed
}

func (f Feed) NewArticles() (a []content.Article) {
	if f.Err() != nil {
		return
	}

	return f.newArticles
}

func (f *Feed) Users() (u []content.User) {
	if f.Err() != nil {
		return
	}

	id := f.Info().Id
	f.logger.Infof("Getting users for feed %d\n", id)

	var in []info.User
	if err := f.db.Select(&in, f.db.SQL("get_feed_users"), id); err != nil {
		f.Err(err)
		return
	}

	u = make([]content.User, len(in))
	for i := range in {
		u[i] = f.Repo().User()
		u[i].Info(in[i])

		if u[i].Err() != nil {
			f.Err(u[i].Err())
			return
		}
	}

	return
}

func (f *Feed) Update() {
	if f.Err() != nil {
		return
	}

	i := f.Info()
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

		i.Id = info.FeedId(id)

		f.Info(i)
	}

	articles := f.updateFeedArticles(tx, f.ParsedArticles())

	if f.Err() != nil {
		return
	}

	f.newArticles = articles

	tx.Commit()
}

func (f *Feed) Delete() {
	if f.Err() != nil {
		return
	}

	id := f.Info().Id
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
	if f.Err() != nil {
		return
	}

	id := f.Info().Id
	f.logger.Infof("Getting all feed %d articles\n", id)

	var info []info.Article
	if err := f.db.Select(&info, f.db.SQL("get_all_feed_articles"), id); err != nil {
		f.Err(err)
		return
	}

	a = make([]content.Article, len(info))
	for i := range info {
		a[i] = f.Repo().Article()
		a[i].Info(info[i])
	}

	return
}

func (f *Feed) LatestArticles() (a []content.Article) {
	if f.Err() != nil {
		return
	}

	id := f.Info().Id
	f.logger.Infof("Getting latest feed %d articles\n", id)

	var info []info.Article
	if err := f.db.Select(&info, f.db.SQL("get_latest_feed_articles"), id); err != nil {
		f.Err(err)
		return
	}

	a = make([]content.Article, len(info))
	for i := range info {
		a[i] = f.Repo().Article()
		a[i].Info(info[i])
	}

	return
}

func (f *Feed) AddArticles(articles []content.Article) {
	if f.Err() != nil {
		return
	}

	id := f.Info().Id
	f.logger.Infof("Adding articles to feed %d\n", id)

	tx, err := f.db.Begin()
	if err != nil {
		f.Err(err)
		return
	}
	defer tx.Rollback()

	newArticles := f.updateFeedArticles(tx, articles)

	if f.Err() != nil {
		return
	}

	tx.Commit()

	f.newArticles = newArticles
}

func (f *Feed) Subscription() (s content.Subscription) {
	if f.Err() != nil {
		return
	}

	id := f.Info().Id
	f.logger.Infof("Getting subcription for feed %d\n", id)

	var in info.Subscription
	if err := f.db.Get(&in, f.db.SQL("get_hubbub_subscription"), id); err != nil && err != dsql.ErrNoRows {
		f.Err(err)
		return
	}

	s = f.Repo().Subscription()
	s.Info(in)

	return
}

func (f *Feed) updateFeedArticles(tx *db.Tx, articles []content.Article) (a []content.Article) {
	if f.Err() != nil || len(articles) == 0 {
		return
	}

	for i := range articles {
		if err := articles[i].Validate(); err != nil {
			f.Err(err)
			return
		}

		id := f.Info().Id
		in := articles[i].Info()
		in.FeedId = id

		var sql string
		args := []interface{}{in.Title, in.Description, in.Date, id}

		if in.Guid.Valid {
			sql = f.db.SQL("update_feed_article_with_guid")
			args = append(args, in.Guid)
		} else {
			sql = f.db.SQL("update_feed_article")
			args = append(args, in.Link)
		}

		stmt, err := tx.Preparex(sql)
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

		if num, err := res.RowsAffected(); err != nil && err == dsql.ErrNoRows || num == 0 {
			a = append(a, articles[i])

			id, err := f.db.CreateWithId(tx, "create_feed_article", id, in.Link, in.Guid,
				in.Title, in.Description, in.Date)

			if err != nil {
				f.Err(err)
				return
			}

			in.Id = info.ArticleId(id)
			articles[i].Info(in)
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
	if uf.Err() != nil {
		return
	}

	id := uf.Info().Id
	login := uf.User().Info().Login
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
	if uf.Err() != nil {
		return
	}

	id := uf.Info().Id
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
	if uf.Err() != nil {
		return
	}

	id := uf.Info().Id
	uf.logger.Infof("Getting unread articles for feed %d\n", id)

	articles := uf.getArticles("ar.article_id IS NULL", "", paging...)
	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
	}

	return
}

func (uf *UserFeed) ReadBefore(date time.Time, read bool) {
	if uf.Err() != nil {
		return
	}

	id := uf.Info().Id
	login := uf.User().Info().Login
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
	if uf.Err() != nil {
		return
	}

	u := uf.User()
	id := uf.Info().Id
	login := u.Info().Login
	uf.logger.Infof("Getting scored articles for user %s feed %d between %v and %v\n", login, id, from, to)

	order := ""
	if uf.Order() == info.DescendingOrder {
		order = " DESC"
	}
	ua := getArticles(u, uf.db, uf.logger, uf, "asco.score",
		"INNER JOIN articles_scores asco ON a.id = asco.article_id",
		"uf.feed_id = $2 AND a.date > $3 AND a.date <= $4", "asco.score"+order,
		[]interface{}{id, from, to}, paging...)

	sa = make([]content.ScoredArticle, len(ua))
	for i := range ua {
		sa[i] = uf.Repo().ScoredArticle()
		sa[i].Info(ua[i].Info())
	}

	return
}

func (uf *UserFeed) getArticles(where, order string, paging ...int) (ua []content.UserArticle) {
	if uf.Err() != nil {
		return
	}

	if where == "" {
		where = "uf.feed_id = $2"
	} else {
		where = "uf.feed_id = $2 AND " + where
	}

	u := uf.User()
	ua = getArticles(u, uf.db, uf.logger, uf, "", "", where, order, []interface{}{uf.Info().Id}, paging...)

	if u.Err() != nil {
		uf.Err(u.Err())
	}

	return
}

func (tf *TaggedFeed) AddTags(tags ...content.Tag) {
	if tf.Err() != nil || len(tags) == 0 {
		return
	}

	id := tf.Info().Id
	login := tf.User().Info().Login
	tf.logger.Infof("Adding tags for feed %d\n", id)

	tx, err := tf.db.Begin()
	if err != nil {
		tf.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(tf.db.SQL("create_user_feed_tag"))
	if err != nil {
		tf.Err(err)
		return
	}
	defer stmt.Close()

	existing := tf.Tags()
	existingMap := make(map[info.TagValue]bool)

	for i := range existing {
		existingMap[existing[i].Value()] = true
	}

	for i := range tags {
		_, err = stmt.Exec(login, id, tags[i].String())
		if err != nil {
			tf.Err(err)
			return
		}

		if !existingMap[tags[i].Value()] {
			existing = append(existing, tags[i])
		}
	}

	tf.Tags(existing)

	tx.Commit()
}

func (tf *TaggedFeed) DeleteAllTags() {
	if tf.Err() != nil {
		return
	}

	id := tf.Info().Id
	login := tf.User().Info().Login
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
		tf.Err(err)
		return
	}

	tf.Tags([]content.Tag{})

	tx.Commit()
}
