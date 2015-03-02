package base

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/urandom/readeef/content/data"
)

type Subscription struct {
	Error
	RepoRelated

	data           data.Subscription
	callbackPrefix string
}

func (s Subscription) String() string {
	return fmt.Sprintf("Subscription for %s\n", s.data.Link)
}

func (s *Subscription) Data(d ...data.Subscription) data.Subscription {
	if s.HasErr() {
		return data.Subscription{}
	}

	if len(d) > 0 {
		s.data = d[0]
	}

	return s.data
}

func (s *Subscription) Validate() error {
	if s.data.Link == "" {
		return ValidationError{errors.New("No subscription link")}
	}

	if u, err := url.Parse(s.data.Link); err != nil || !u.IsAbs() {
		return ValidationError{errors.New("Invalid subscription link")}
	}

	if s.data.FeedId == 0 {
		return ValidationError{errors.New("Invalid feed id")}
	}

	return nil
}

func (s Subscription) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.data)
}
