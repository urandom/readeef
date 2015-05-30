package base

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

type Subscription struct {
	Error
	RepoRelated

	data           data.Subscription
	callbackPrefix string
}

func (s Subscription) String() string {
	return "Subscription for " + s.data.Link
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
		return content.NewValidationError(errors.New("No subscription link"))
	}

	if u, err := url.Parse(s.data.Link); err != nil || !u.IsAbs() {
		return content.NewValidationError(errors.New("Invalid subscription link"))
	}

	if s.data.FeedId == 0 {
		return content.NewValidationError(errors.New("Invalid feed id"))
	}

	return nil
}

func (s Subscription) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(s.data)

	if err == nil {
		return b, nil
	} else {
		return []byte{}, fmt.Errorf("Error marshaling subscription data for %s: %v", s, err)
	}
}
