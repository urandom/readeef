package parser

import (
	"reflect"
	"testing"
	"time"
)

func TestParseAtom(t *testing.T) {
	tests := []struct {
		name    string
		b       []byte
		want    Feed
		wantErr bool
	}{
		{"single", []byte(singleAtomXML), singleAtomFeed, false},
		{"single publish date", []byte(singlePubAtomXML), singleAtomFeed, false},
		{"single no date", []byte(singleNoDateAtomXML), singleNoDateAtomFeed, false},
		{"multi last no date", []byte(multiLastNoDateAtomXML), multiLastNoDateAtomFeed, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAtom(tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAtom() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseAtom() = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	singleAtomFeed = Feed{
		Title:    "Example Feed",
		SiteLink: "http://example.org/",
		Articles: []Article{
			{
				Title:       "Atom-Powered Robots Run Amok",
				Link:        "http://example.org/2003/12/13/atom03",
				Guid:        "urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a",
				Description: "Some text.",
				Date:        time.Date(2003, time.December, 13, 18, 30, 02, 0, time.UTC),
			},
		},
	}

	singleNoDateAtomFeed = Feed{
		Title:    "Example Feed",
		SiteLink: "http://example.org/",
		Articles: []Article{
			{
				Title:       "Atom-Powered Robots Run Amok",
				Link:        "http://example.org/2003/12/13/atom03",
				Guid:        "urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a",
				Description: "Some text.",
				Date:        time.Unix(0, 0),
			},
		},
	}

	multiLastNoDateAtomFeed = Feed{
		Title:    "Example Feed",
		SiteLink: "http://example.org/",
		Articles: []Article{
			{
				Title:       "Atom-Powered Robots Run Amok",
				Link:        "http://example.org/2003/12/13/atom03",
				Guid:        "urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a",
				Description: "Some text.",
				Date:        time.Date(2003, time.December, 13, 18, 30, 02, 0, time.UTC),
			},
			{
				Title:       "Atom-Powered Robots Run Amok 2",
				Link:        "http://example.org/2003/12/13/atom03 2",
				Guid:        "urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a 2",
				Description: "Some text. 2",
				Date:        time.Date(2003, time.December, 13, 18, 30, 03, 0, time.UTC),
			},
		},
	}
)

const (
	singleAtomXML = `
<feed xmlns="http://www.w3.org/2005/Atom" updated="2003-12-13T18:30:02Z">
	<title>Example Feed</title>
	<id>urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6</id>
	<link href="http://example.org/"></link>
	<author>
		<name>John Doe</name>
		<uri></uri>
		<email></email>
	</author>
	<entry>
		<title>Atom-Powered Robots Run Amok</title>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<link href="http://example.org/2003/12/13/atom03"></link>
		<updated>2003-12-13T18:30:02Z</updated>
		<author><name></name><uri></uri><email></email></author>
		<summary>Some text.</summary>
	</entry>
</feed>
`
	singlePubAtomXML = `
<feed xmlns="http://www.w3.org/2005/Atom" updated="2003-12-13T18:30:02Z">
	<title>Example Feed</title>
	<id>urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6</id>
	<link href="http://example.org/"></link>
	<author>
		<name>John Doe</name>
		<uri></uri>
		<email></email>
	</author>
	<entry>
		<title>Atom-Powered Robots Run Amok</title>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<link href="http://example.org/2003/12/13/atom03"></link>
		<published>2003-12-13T18:30:02Z</published>
		<author><name></name><uri></uri><email></email></author>
		<content>Some text.</content>
	</entry>
</feed>
`

	singleNoDateAtomXML = `
<feed xmlns="http://www.w3.org/2005/Atom" updated="2003-12-13T18:30:02Z">
	<title>Example Feed</title>
	<id>urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6</id>
	<link href="http://example.org/"></link>
	<author>
		<name>John Doe</name>
		<uri></uri>
		<email></email>
	</author>
	<entry>
		<title>Atom-Powered Robots Run Amok</title>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<link href="http://example.org/2003/12/13/atom03"></link>
		<author><name></name><uri></uri><email></email></author>
		<summary>Some text.</summary>
	</entry>
</feed>
`

	multiLastNoDateAtomXML = `
<feed xmlns="http://www.w3.org/2005/Atom" updated="2003-12-13T18:30:02Z">
	<title>Example Feed</title>
	<id>urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6</id>
	<link href="http://example.org/"></link>
	<author>
		<name>John Doe</name>
		<uri></uri>
		<email></email>
	</author>
	<entry>
		<title>Atom-Powered Robots Run Amok</title>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<link href="http://example.org/2003/12/13/atom03"></link>
		<updated>2003-12-13T18:30:02Z</updated>
		<author><name></name><uri></uri><email></email></author>
		<summary>Some text.</summary>
	</entry>
	<entry>
		<title>Atom-Powered Robots Run Amok 2</title>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a 2</id>
		<link href="http://example.org/2003/12/13/atom03 2"></link>
		<author><name></name><uri></uri><email></email></author>
		<summary>Some text. 2</summary>
	</entry>
</feed>
`
)
