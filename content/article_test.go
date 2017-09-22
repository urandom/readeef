package content_test

import (
	"testing"

	"github.com/urandom/readeef/content"
)

func TestArticle_Validate(t *testing.T) {
	type fields struct {
		ID     content.ArticleID
		FeedID content.FeedID
		Link   string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"valid", fields{ID: 1, FeedID: 1, Link: "http://sugr.org"}, false},
		{"link not absolute", fields{ID: 1, FeedID: 1, Link: "sugr.org"}, true},
		{"no link", fields{ID: 1, FeedID: 1}, true},
		{"no feed id", fields{ID: 1, Link: "http://sugr.org"}, true},
		{"no id", fields{FeedID: 1, Link: "http://sugr.org"}, true},
		{"nothing", fields{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := content.Article{
				ID:     tt.fields.ID,
				FeedID: tt.fields.FeedID,
				Link:   tt.fields.Link,
			}
			if err := a.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Article.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
