package data

import (
	"database/sql/driver"
	"fmt"
)

type TagValue string
type TagId int64

type Tag struct {
	Id     TagId
	FeedId FeedId   `db:"feed_id"`
	Value  TagValue `db:"tag"`
}

func (val *TagValue) Scan(src interface{}) error {
	switch t := src.(type) {
	case string:
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
