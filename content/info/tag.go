package info

import "fmt"

type TagValue string

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
