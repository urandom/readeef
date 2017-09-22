package content_test

import (
	"testing"

	"github.com/urandom/readeef/content"
)

func TestScores_Validate(t *testing.T) {
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
			s := content.Scores{
				ArticleID: tt.ArticleID,
			}
			if err := s.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Scores.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
