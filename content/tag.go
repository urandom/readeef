package content

import (
	"database/sql/driver"
	"errors"
	"fmt"
)

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

func (id *TagID) Scan(src interface{}) error {
	asInt, ok := src.(int64)
	if !ok {
		return fmt.Errorf("Scan source '%#v' (%T) was not of type int64 (TagId)", src, src)
	}

	*id = TagID(asInt)

	return nil
}

func (id TagID) Value() (driver.Value, error) {
	return int64(id), nil
}

func (val *TagValue) Scan(src interface{}) error {
	switch t := src.(type) {
	case string:
		*val = TagValue(t)
	case []byte:
		*val = TagValue(t)
	default:
		return fmt.Errorf("Scan source '%#v' (%T) was not of type string (TagValue)", src, src)
	}

	return nil
}

func (val TagValue) Value() (driver.Value, error) {
	return string(val), nil
}
