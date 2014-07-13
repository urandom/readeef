package readeef

import "time"

type Feed interface {
	Title() string
	Description() string
	Link() string
	HubLink() string
	Image() Image
	Articles() []Article
}

type Article interface {
	Id() string
	Title() string
	Description() string
	Link() string
	Date() time.Time
}

type Image interface {
	Title() string
	Url() string
	Width() int
	Height() int
}
