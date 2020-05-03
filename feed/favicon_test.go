package feed

import (
	"testing"
)

func TestFavicon(t *testing.T) {
	tests := []struct {
		name    string
		site    string
		want    bool
		wantCT  string
		wantErr bool
	}{
		{name: "no url", wantErr: true},
		{name: "broken url", site: "https://example-nonexistent.com", wantErr: true},
		{name: "wikipedia", site: "https://en.wikipedia.org", want: true, wantCT: "image/vnd.microsoft.icon"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ct, err := Favicon(tt.site)
			if (err != nil) != tt.wantErr {
				t.Errorf("Favicon() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (len(got) > 0) != tt.want {
				t.Errorf("Favicon() = %v, want %v", got, tt.want)
			}

			if ct != tt.wantCT {
				t.Errorf("Favicon()  content-type = %q, want %q", ct, tt.wantCT)
			}
		})
	}
}
