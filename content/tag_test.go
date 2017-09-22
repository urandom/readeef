package content_test

import (
	"testing"

	"github.com/urandom/readeef/content"
)

func TestTag_Validate(t *testing.T) {
	tests := []struct {
		name    string
		Value   content.TagValue
		wantErr bool
	}{
		{"valid", "tag", false},
		{"invalid", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := content.Tag{
				Value: tt.Value,
			}
			if err := tag.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Tag.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
