package content_test

import (
	"testing"

	"github.com/urandom/readeef/content"
)

func TestExtract_Validate(t *testing.T) {
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
			e := content.Extract{
				ArticleID: tt.ArticleID,
			}
			if err := e.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Extract.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
