package parser

import "testing"

func TestParseFeed(t *testing.T) {
	_, err := ParseFeed([]byte(singleAtomXML), ParseRss2, ParseAtom, ParseRss1)

	if err != nil {
		t.Fatal(err)
	}

	_, err = ParseFeed([]byte(singleRss1XML), ParseRss2, ParseAtom, ParseRss1)

	if err != nil {
		t.Fatal(err)
	}

	_, err = ParseFeed([]byte(singleRss2XML), ParseRss2, ParseAtom, ParseRss1)

	if err != nil {
		t.Fatal(err)
	}

	_, err = ParseFeed([]byte(singleRss1XML), ParseRss2, ParseAtom)

	if err == nil {
		t.Fatalf("Expected an error\n")
	}

	_, err = ParseFeed([]byte(singleRss2XML), ParseRss1, ParseAtom)

	if err == nil {
		t.Fatalf("Expected an error\n")
	}
}
