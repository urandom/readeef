package base

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/info"
)

type Feed struct {
	ArticleSorting
	Error

	info info.Feed
}

type UserFeed struct {
	Feed
	ArticleSearch

	user content.User
}

type TaggedFeed struct {
	UserFeed
	tags []content.Tag
}

func NewUserFeed(user content.User) UserFeed {
	return UserFeed{user: user}
}

func (f Feed) String() string {
	return f.info.Title + " " + strconv.FormatInt(int64(f.info.Id), 10)
}

func (f *Feed) Set(info info.Feed) content.Feed {
	if f.Err() != nil {
		return f
	}

	f.info = info

	return f
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

func (f Feed) AddArticles([]content.Article) content.Feed {
	panic("Not implemented")
}

func (f Feed) AllArticles() []content.Article {
	panic("Not implemented")
}

func (f Feed) Delete() content.Feed {
	panic("Not implemented")
}

func (f Feed) LatestArticles() []content.Article {
	panic("Not implemented")
}

func (f Feed) Subscription() content.Subscription {
	panic("Not implemented")
}

func (f Feed) NewArticles() []content.Article {
	panic("Not implemented")
}

func (f Feed) Update(i info.Feed) content.Feed {
	panic("Not implemented")
}

func (f UserFeed) User() content.User {
	return f.user
}

func (f UserFeed) Validate() error {
	err := f.Feed.Validate()
	if err != nil {
		return err
	}

	if f.user.Info().Login == "" {
		return ValidationError{errors.New("UserFeed has no user")}
	}

	return nil
}

func (uf UserFeed) Detach() content.UserFeed {
	panic("Not implemented")
}

func (uf UserFeed) Articles(desc bool, paging ...int) []content.UserArticle {
	panic("Not implemented")
}

func (uf UserFeed) UnreadArticles(desc bool, paging ...int) []content.UserArticle {
	panic("Not implemented")
}

func (uf UserFeed) ReadBefore(date time.Time, read bool) content.UserFeed {
	panic("Not implemented")
}

func (uf UserFeed) ScoredArticles(from, to time.Time, paging ...int) []content.ScoredArticle {
	panic("Not implemented")
}

func (uf UserFeed) Users() []content.User {
	panic("Not implemented")
}

func (tf TaggedFeed) Tags() []content.Tag {
	return tf.tags
}

func (tf *TaggedFeed) SetTags(tags ...content.Tag) content.TaggedFeed {
	tf.tags = tags
	return tf
}

func (tf *TaggedFeed) AddTags(tags ...content.Tag) content.TaggedFeed {
	panic("Not implemented")
	return tf
}

func (tf TaggedFeed) DeleteAllTags() content.TaggedFeed {
	panic("Not implemented")
}
