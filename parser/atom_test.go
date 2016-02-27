package parser

import "testing"

var atomXml = `` +
	`<feed xmlns="http://www.w3.org/2005/Atom" updated="2003-12-13T18:30:02Z">` +
	`<title>Example Feed</title>` +
	`<id>urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6</id>` +
	`<link href="http://example.org/"></link>` +
	`<author><name>John Doe</name><uri></uri><email></email></author>` +
	`<entry>` +
	`<title>Atom-Powered Robots Run Amok</title>` +
	`<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>` +
	`<link href="http://example.org/2003/12/13/atom03"></link>` +
	`<updated>2003-12-13T18:30:02Z</updated>` +
	`<author><name></name><uri></uri><email></email></author>` +
	`<summary>Some text.</summary>` +
	`</entry>` +
	`</feed>`

func TestAtomParse(t *testing.T) {
	f, err := ParseAtom([]byte(""))
	if err == nil {
		t.Fatalf("Expected an error, got none, feed: '%v'\n", f)
	}

	f, err = ParseAtom([]byte(atomXml))
	if err != nil {
		t.Fatal(err)
	}

	if f.Title != "Example Feed" {
		t.Fatalf("Unexpected feed title: '%s'\n", f.Title)
	}

	if f.Description != "" {
		t.Fatalf("Unexpected feed description: '%s'\n", f.Description)
	}

	if f.SiteLink != "http://example.org/" {
		t.Fatalf("Unexpected feed link: '%s'\n", f.SiteLink)
	}

	if f.HubLink != "" {
		t.Fatalf("Unexpected feed hub link: '%s'\n", f.HubLink)
	}

	if f.Image != (Image{}) {
		t.Fatalf("Unexpected feed image: '%v'\n", f.Image)
	}

	if 1 != len(f.Articles) {
		t.Fatalf("Unexpected number of feed articles: '%d'\n", len(f.Articles))
	}

	a := f.Articles[0]
	expectedStr := "urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a"
	if a.Guid != expectedStr {
		t.Fatalf("Expected %s as id, got '%s'\n", expectedStr, a.Guid)
	}

	if a.Title != "Atom-Powered Robots Run Amok" {
		t.Fatalf("Unexpected article title: '%v'\n", a.Title)
	}

	if a.Description != "Some text." {
		t.Fatalf("Unexpected article description: '%v'\n", a.Description)
	}

	if a.Link != "http://example.org/2003/12/13/atom03" {
		t.Fatalf("Unexpected article link: '%v'\n", a.Link)
	}

	d, _ := parseDate("2003-12-13T18:30:02Z")

	if !a.Date.Equal(d) {
		t.Fatalf("Unexpected article date: '%s'\n", a.Date)
	}
}
