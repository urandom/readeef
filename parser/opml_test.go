package parser

import (
	"reflect"
	"testing"
)

func TestParseOpml(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		want    Opml
		wantErr bool
	}{
		{"single", []byte(singleOmplXML), singleOpml, false},
		{"single url", []byte(singleUrlOmplXML), singleTagsOpml, false},
		{"deep", []byte(deepOpmlXML), deepOpml, false},
		{"error", []byte("<foobar/>"), Opml{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseOpml(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseOpml() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseOpml() = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	singleOpml = Opml{
		Feeds: []OpmlFeed{
			{
				Title: "Item 1 text",
				URL:   "http://www.item1.com/rss",
				Tags:  []string{},
			},
		},
	}
	singleTagsOpml = Opml{
		Feeds: []OpmlFeed{
			{
				Title: "Item 1 text",
				URL:   "http://www.item1.com/rss",
				Tags:  []string{"cat1", "cat2"},
			},
		},
	}
	deepOpml = Opml{
		Feeds: []OpmlFeed{
			{
				Title: "Item 1 text",
				URL:   "http://www.item1.com/rss",
				Tags:  []string{"cat1", "cat2"},
			},
			{
				Title: "Item 2 text",
				URL:   "http://www.item2.com/rss",
				Tags:  []string{"cat1", "cat2"},
			},
		},
	}
)

const (
	singleOmplXML = `
<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.1">
    <head>
        <title>
			OPML title
		</title>
    </head>
    <body>
        <outline type="rss" text="Item 1 text" title="Item 1 title" xmlUrl="http://www.item1.com/rss" htmlUrl="http://www.item1.com"></outline>
    </body>
</opml>
`
	singleUrlOmplXML = `
<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.1">
    <head>
        <title>
			OPML title
		</title>
    </head>
    <body>
        <outline type="rss" text="Item 1 text" title="Item 1 title" url="http://www.item1.com/rss" htmlUrl="http://www.item1.com" category="cat1, cat2"></outline>
    </body>
</opml>
`

	deepOpmlXML = `
<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.1">
    <head>
        <title>
			OPML title
		</title>
    </head>
    <body>
		<outline text="cat1, cat2">
			<outline type="rss" text="Item 1 text" title="Item 1 title" url="http://www.item1.com/rss" htmlUrl="http://www.item1.com"></outline>
			<outline type="rss" text="Item 2 text" title="Item 2 title" url="http://www.item2.com/rss" htmlUrl="http://www.item2.com"></outline>
		</outline>
    </body>
</opml>
`
)
