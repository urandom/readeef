package base

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

type Tag struct {
	ArticleSorting
	Error
	UserRelated
	RepoRelated

	data data.Tag
}

func (t Tag) String() string {
	return string(t.data.Value)
}

func (t *Tag) Data(d ...data.Tag) data.Tag {
	if t.HasErr() {
		return data.Tag{}
	}

	if len(d) > 0 {
		t.data = d[0]
	}

	return t.data
}

func (t Tag) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(t.data.Value)

	if err == nil {
		return b, nil
	} else {
		return []byte{}, fmt.Errorf("Error marshaling tag %s: %v", t, err)
	}
}

func (t *Tag) Validate() error {
	if t.data.Value == "" {
		return content.NewValidationError(errors.New("Tag has no value"))
	}

	if t.user == nil || t.user.Data().Login == "" {
		return content.NewValidationError(errors.New("Tag has no user"))
	}

	return nil
}
