package content_test

import (
	"testing"

	"github.com/urandom/readeef/content"
)

func TestSubscription_Validate(t *testing.T) {
	type fields struct {
		Link   string
		FeedID content.FeedID
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"valid", fields{FeedID: 1, Link: "http://sugr.org"}, false},
		{"link not absolute", fields{FeedID: 1, Link: "sugr.org"}, true},
		{"no link", fields{FeedID: 1}, true},
		{"no feed id", fields{Link: "http://sugr.org"}, true},
		{"nothing", fields{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := content.Subscription{
				Link:   tt.fields.Link,
				FeedID: tt.fields.FeedID,
			}
			if err := s.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Subscription.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
