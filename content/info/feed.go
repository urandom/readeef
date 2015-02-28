package info

import "time"

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
