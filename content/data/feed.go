package data

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type FeedId int64

type Feed struct {
	Id             FeedId
	Title          string
	Description    string
	Link           string
	SiteLink       string          `db:"site_link",json:"-"`
	HubLink        string          `db:"hub_link",json:"-"`
	UpdateError    string          `db:"update_error"`
	SubscribeError string          `db:"subscribe_error"`
	TTL            time.Duration   `json:"-"`
	SkipHours      map[int]bool    `json:"-"`
	SkipDays       map[string]bool `json:"-"`
}

func (id *FeedId) Scan(src interface{}) error {
	asInt, ok := src.(int64)
	if !ok {
		return fmt.Errorf("Scan source '%#v' (%T) was not of type int64 (FeedId)", src, src)
	}

	(*id) = FeedId(asInt)

	return nil
}

func (id FeedId) Value() (driver.Value, error) {
	return int64(id), nil
}
