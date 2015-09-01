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

	value data.TagValue
}

func (t Tag) String() string {
	return string(t.value)
}

func (t *Tag) Value(val ...data.TagValue) data.TagValue {
	if t.HasErr() {
		return ""
	}

	if len(val) > 0 {
		t.value = val[0]
	}

	return t.value
}

func (t Tag) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(t.value)

	if err == nil {
		return b, nil
	} else {
		return []byte{}, fmt.Errorf("Error marshaling tag %s: %v", t, err)
	}
}

func (t *Tag) Validate() error {
	if t.value == "" {
		return content.NewValidationError(errors.New("Tag has no value"))
	}

	if t.user == nil || t.user.Data().Login == "" {
		return content.NewValidationError(errors.New("Tag has no user"))
	}

	return nil
}
