package parser

import "time"

type Feed struct {
	Title       string
	Description string
	Link        string
	HubLink     string
	Image       Image
	Articles    []Article
	TTL         time.Duration
	SkipHours   map[int]bool
	SkipDays    map[string]bool
}

type Article struct {
	Id          string
	Title       string
	Description string
	Link        string
	Date        time.Time
}

type Image struct {
	Title  string
	Url    string
	Width  int
	Height int
}
