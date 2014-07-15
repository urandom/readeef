package readeef

const (
	get_user_feeds = `
SELECT f.id, f.title, f.description, f.link, f.hub_link
FROM feeds f, users_feeds uf
WHERE f.id = uf.feed_id
	AND uf.user_login = ?
`

	get_feed_articles = `
SELECT uf.feed_id, a.id, a.title, a.description, a.link, a.date,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = articles.feed_id AND uf.feed_id = ? AND uf.user_login = ?
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id
LIMIT ?
OFFSET ?
`

	get_user_favorite_articles = `
SELECT uf.feed_id, a.id, a.title, a.description, a.link, a.date
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = articles.feed_id AND uf.user_login = ?
INNER JOIN users_articles_fav af
	ON a.id = af.article_id
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id
LIMIT ?
OFFSET ?
`
)

func (db DB) GetUserFeeds(u User) ([]Feed, error) {
	var feeds []Feed

	if err := db.Select(&feeds, get_user_feeds, u.Login); err != nil {
		return feeds, err
	}

	for _, f := range feeds {
		f.User = u
	}

	return feeds, nil
}

func (db DB) GetFeedArticles(f Feed, paging ...int) (Feed, error) {
	var articles []Article

	limit, offset := pagingLimit(paging)

	if err := db.Select(&articles, get_feed_articles, f.Id, f.User.Login, limit, offset); err != nil {
		return f, err
	}

	f.Articles = articles

	return f, nil
}

func (db DB) GetUserFavoriteArticles(u User, paging ...int) ([]Article, error) {
	var articles []Article
	limit, offset := pagingLimit(paging)

	if err := db.Select(&articles, get_user_favorite_articles, u.Login, limit, offset); err != nil {
		return articles, err
	}

	return articles, nil
}

func pagingLimit(paging []int) (int, int) {
	limit := 20
	offset := 0

	if len(paging) > 0 {
		limit = paging[0]
		if len(paging) > 1 {
			offset = paging[1]
		}
	}

	return limit, offset
}
