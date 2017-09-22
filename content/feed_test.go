package content_test

import (
	"testing"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/parser"
)

func TestFeed_Validate(t *testing.T) {
	type fields struct {
		ID   content.FeedID
		Link string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"valid", fields{ID: 1, Link: "http://sugr.org"}, false},
		{"link not absolute", fields{ID: 1, Link: "sugr.org"}, true},
		{"no link", fields{ID: 1}, true},
		{"no id", fields{Link: "http://sugr.org"}, true},
		{"nothing", fields{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := content.Feed{
				ID:   tt.fields.ID,
				Link: tt.fields.Link,
			}
			if err := f.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Feed.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFeed_Refresh(t *testing.T) {
	tests := []struct {
		name   string
		feed   content.Feed
		parsed parser.Feed
	}{
		{"simple feed", content.Feed{}, parser.Feed{Title: "Title", Description: "Description", SiteLink: "http://sugr.org"}},
		{"simple feed 2", content.Feed{Title: "Diff", Description: "Diff"}, parser.Feed{Title: "Title", Description: "Description", SiteLink: "http://sugr.org", HubLink: "http://hub.sugr.org"}},
		{"with articles", content.Feed{}, parser.Feed{Title: "Title", Articles: []parser.Article{
			{Title: "Title 1"},
			{Title: "Title 2"},
		}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.feed
			f.Refresh(tt.parsed)

			if f.Title != tt.parsed.Title {
				t.Errorf("Feed.Refresh() title want = %v, got %v", tt.parsed.Title, f.Title)
			}

			if f.Description != tt.parsed.Description {
				t.Errorf("Feed.Refresh() description want = %v, got %v", tt.parsed.Description, f.Description)
			}

			if f.SiteLink != tt.parsed.SiteLink {
				t.Errorf("Feed.Refresh() siteLink want = %v, got %v", tt.parsed.SiteLink, f.SiteLink)
			}

			if f.HubLink != tt.parsed.HubLink {
				t.Errorf("Feed.Refresh() hubLink want = %v, got %v", tt.parsed.HubLink, f.HubLink)
			}

			if len(f.ParsedArticles()) != len(tt.parsed.Articles) {
				t.Errorf("Feed.Refresh() articles want = %v, got %v", len(tt.parsed.Articles), len(f.ParsedArticles()))
			}

			for i, a := range f.ParsedArticles() {
				if a.Title != tt.parsed.Articles[i].Title {
					t.Errorf("Feed.Refresh() article %d title want = %v, got %v", i, tt.parsed.Articles[i].Title, a.Title)
				}
			}
		})
	}
}
