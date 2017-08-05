package content

import "errors"

type TagID int64
type TagValue string

type Tag struct {
	ID    TagID    `json:"id"`
	Value TagValue `json:"value"`
}

func (t Tag) Validate() error {
	if t.Value == "" {
		return NewValidationError(errors.New("Tag has no value"))
	}

	return nil
}

func (t Tag) String() string {
	return string(t.Value)
}
