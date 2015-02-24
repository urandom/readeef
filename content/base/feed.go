package base

import (
	"errors"
	"net/url"
	"strconv"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/info"
	"github.com/urandom/readeef/parser"
)

type Feed struct {
	ArticleSorting
	Error

	info           info.Feed
	parsedArticles []content.Article
}

type UserFeed struct {
	ArticleSearch

	user content.User
}

type TaggedFeed struct {
	tags []content.Tag
}

func NewUserFeed(user content.User) UserFeed {
	return UserFeed{user: user}
}

func (f Feed) String() string {
	return f.info.Title + " " + strconv.FormatInt(int64(f.info.Id), 10)
}

func (f *Feed) Set(info info.Feed) {
	if f.Err() != nil {
		return
	}

	f.info = info
}

func (f Feed) Info() info.Feed {
	return f.info
}

func (f Feed) Validate() error {
	if u, err := url.Parse(f.info.Link); err != nil || !u.IsAbs() {
		return ValidationError{errors.New("Feed has no link")}
	}

	return nil
}

func (f *Feed) Refresh(pf parser.Feed) {
	if f.Err() != nil {
		return
	}

	in := f.Info()

	in.Title = pf.Title
	in.Description = pf.Description
	in.SiteLink = pf.SiteLink
	in.HubLink = pf.HubLink

	newArticles := make([]Article, len(pf.Articles))

	for i := range pf.Articles {
		ai := info.Article{
			Title:       pf.Articles[i].Title,
			Description: pf.Articles[i].Description,
			Link:        pf.Articles[i].Link,
			Date:        pf.Articles[i].Date,
		}
		ai.FeedId = in.Id

		if pf.Articles[i].Guid != "" {
			ai.Guid.Valid = true
			ai.Guid.String = pf.Articles[i].Guid
		}

		a := Article{info: ai}

		newArticles[i] = a
	}

	f.Set(in)
}

func (f *Feed) ParsedArticles() (a []content.Article) {
	if f.Err() != nil {
		return
	}

	return f.parsedArticles
}

func (f UserFeed) User() content.User {
	return f.user
}

func (f UserFeed) Validate() error {
	if f.user.Info().Login == "" {
		return ValidationError{errors.New("UserFeed has no user")}
	}

	return nil
}

func (tf TaggedFeed) Tags() []content.Tag {
	return tf.tags
}

func (tf *TaggedFeed) SetTags(tags []content.Tag) {
	tf.tags = tags
}
