package base

import (
	"errors"
	"net/url"
	"strconv"

	"github.com/urandom/readeef/content/info"
)

type Feed struct {
	ArticleSorting
	Error

	info info.Feed
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
