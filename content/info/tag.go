package info

import "fmt"

type TagValue string

func (val *TagValue) Scan(src interface{}) error {
	asString, ok := src.(string)
	if !ok {
		return fmt.Errorf("Scan source '%T' was not of type string", src)
	}

	*val = TagValue(asString)

	return nil
}
