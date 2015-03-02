package base

import (
	"encoding/json"
	"errors"

	"github.com/urandom/readeef/content/data"
)

type Tag struct {
	ArticleSorting
	ArticleSearch
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
	return json.Marshal(t.value)
}

func (t *Tag) Validate() error {
	if t.value == "" {
		return ValidationError{errors.New("Tag has no value")}
	}

	return nil
}
