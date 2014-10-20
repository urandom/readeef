package readeef

import (
	"database/sql"
	"errors"
	"net/url"

	"github.com/urandom/readeef/parser"
)

type Feed struct {
	parser.Feed

	User           User
	Id             int64
	Link           string
	SiteLink       string `db:"site_link"`
	HubLink        string `db:"hub_link"`
	UpdateError    string `db:"update_error"`
	SubscribeError string `db:"subscribe_error"`
	Articles       []Article
	Tags           []string

	lastUpdatedArticleLinks map[string]bool
}

type Article struct {
	parser.Article

	Id       int64
	Guid     sql.NullString
	FeedId   int64 `db:"feed_id"`
	Read     bool
	Favorite bool
}

type ArticleScores struct {
	ArticleId int64
	Score     int
	Score1    int
	Score2    int
	Score3    int
	Score4    int
	Score5    int
}

var (
	EmptyArticleScores = ArticleScores{}
)

func (f Feed) UpdateFromParsed(pf parser.Feed) Feed {
	f.Feed = pf
	f.HubLink = pf.HubLink
	f.SiteLink = pf.SiteLink

	newArticles := make([]Article, len(pf.Articles))

	for i, pa := range pf.Articles {
		a := Article{Article: pa}
		a.FeedId = f.Id

		if pa.Guid != "" {
			a.Guid.Valid = true
			a.Guid.String = pa.Guid
		}

		newArticles[i] = a
	}

	f.Articles = newArticles

	return f
}

func (f Feed) Validate() error {
	if u, err := url.Parse(f.Link); err != nil || !u.IsAbs() {
		return ValidationError{errors.New("Feed has no link")}
	}

	return nil
}

func (a Article) Validate() error {
	if a.FeedId == 0 {
		return ValidationError{errors.New("Article has no feed id")}
	}

	return nil
}

func (asc ArticleScores) Validate() error {
	if asc.ArticleId == 0 {
		return ValidationError{errors.New("Article scores has no article id")}
	}

	return nil
}

func (asc *ArticleScores) CalculateScore() {
	asc.Score = asc.Score1 + int(0.1*float32(asc.Score2)) + int(0.01*float32(asc.Score3)) + int(0.001*float32(asc.Score4)) + int(0.0001*float32(asc.Score5))
}
