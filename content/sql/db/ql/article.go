package ql

const (
	createUserArticleRead = `
INSERT INTO users_articles_read(user_login, article_id, _login_id) VALUES ($1, $2 $1 + ":" + formatInt($2))`
	createUserArticleFavorite = `
INSERT INTO users_articles_fav(user_login, article_id, _login_id) VALUES ($1, $2, $1 + ":" + formatInt($2))
`
	createArticleScores = `
INSERT INTO articles_scores(article_id, score, score1, score2, score3, score4, score5)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
`
)
