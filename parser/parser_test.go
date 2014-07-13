package parser

import "testing"

func TestParseFeed(t *testing.T) {
	_, err := ParseFeed([]byte(atomXml), ParseRss2, ParseAtom, ParseRss1)

	if err != nil {
		t.Fatal(err)
	}

	_, err = ParseFeed([]byte(rss1Xml), ParseRss2, ParseAtom, ParseRss1)

	if err != nil {
		t.Fatal(err)
	}

	_, err = ParseFeed([]byte(rss2Xml), ParseRss2, ParseAtom, ParseRss1)

	if err != nil {
		t.Fatal(err)
	}

	_, err = ParseFeed([]byte(rss1Xml), ParseRss2, ParseAtom)

	if err == nil {
		t.Fatalf("Expected an error\n")
	}

}
