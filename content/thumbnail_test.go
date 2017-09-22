package content_test

import (
	"testing"

	"github.com/urandom/readeef/content"
)

func TestThumbnail_Validate(t *testing.T) {
	tests := []struct {
		name      string
		ArticleID content.ArticleID
		wantErr   bool
	}{
		{"valid", 1, false},
		{"invalid", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			thumb := content.Thumbnail{
				ArticleID: tt.ArticleID,
			}
			if err := thumb.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Thumbnail.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
