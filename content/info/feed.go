package info

import "time"

type FeedId int64

type Feed struct {
	Id             FeedId
	Title          string
	Description    string
	Link           string
	SiteLink       string `db:"site_link"`
	HubLink        string `db:"hub_link"`
	UpdateError    string `db:"update_error"`
	SubscribeError string `db:"subscribe_error"`
	TTL            time.Duration
	SkipHours      map[int]bool
	SkipDays       map[string]bool
}
