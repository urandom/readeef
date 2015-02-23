package base

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/info"
)

type Subscription struct {
	Error

	info           info.Subscription
	callbackPrefix string
}

func (s Subscription) String() string {
	return fmt.Sprintf("Subscription for %s\n", s.info.Link)
}

func (s *Subscription) Set(info info.Subscription) content.Subscription {
	if s.Err() != nil {
		return s
	}

	s.info = info

	return s
}

func (s Subscription) Info() info.Subscription {
	return s.info
}

func (s *Subscription) Validate() error {
	if u, err := url.Parse(s.info.Link); err != nil || !u.IsAbs() {
		return ValidationError{errors.New("Invalid subscription link")}
	}

	if s.info.FeedId == 0 {
		return ValidationError{errors.New("Invalid feed id")}
	}

	return nil
}

func (s Subscription) Delete() content.Subscription {
	panic("Not implemented")
}

func (s Subscription) Fail(fail bool) content.Subscription {
	panic("Not implemented")
}

func (s Subscription) Update(info info.Subscription) content.Subscription {
	panic("Not implemented")
}
