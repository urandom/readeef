package data

import (
	"database/sql/driver"
	"fmt"
)

type TagId int64
type TagValue string

type Tag struct {
	Id    TagId
	Value TagValue
}

func (id *TagId) Scan(src interface{}) error {
	asInt, ok := src.(int64)
	if !ok {
		return fmt.Errorf("Scan source '%#v' (%T) was not of type int64 (TagId)", src, src)
	}

	*id = TagId(asInt)

	return nil
}

func (id TagId) Value() (driver.Value, error) {
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

func (val TagValue) String() string {
	return string(val)
}
