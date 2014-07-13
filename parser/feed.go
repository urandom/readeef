package parser

import "time"

type feed struct {
	title       string
	description string
	link        string
	hubLink     string
	image       image
	articles    []article
}

func (f feed) Articles() []article {
	return f.articles
}

type article struct {
	id          string
	title       string
	description string
	link        string
	date        time.Time
}

type image struct {
	title  string
	url    string
	width  int
	height int
}

func (f feed) Title() string {
	return f.title
}

func (f feed) Description() string {
	return f.description
}

func (f feed) Link() string {
	return f.link
}

func (f feed) HubLink() string {
	return f.hubLink
}

func (f feed) Image() image {
	return f.image
}

func (i article) Id() string {
	return i.id
}

func (i article) Title() string {
	return i.title
}

func (i article) Description() string {
	return i.description
}

func (i article) Link() string {
	return i.link
}

func (i article) Date() time.Time {
	return i.date
}
func (i image) Title() string {
	return i.title
}

func (i image) Url() string {
	return i.url
}

func (i image) Width() int {
	return i.width
}

func (i image) Height() int {
	return i.height
}
