package api

import (
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"testing"
)

func Test_featuresHandler(t *testing.T) {
	tests := []struct {
		name     string
		features features
		wantErr  bool
	}{
		{"all", features{Search: true, Extractor: true, ProxyHTTP: true, Popularity: true}, false},
		{"some", features{Extractor: true}, false},
	}
	type data struct {
		Features features `json:"features"`
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/features", nil)
			w := httptest.NewRecorder()

			featuresHandler(tt.features).ServeHTTP(w, req)
			got := data{}

			if err := json.Unmarshal(w.Body.Bytes(), &got); (err != nil) != tt.wantErr {
				t.Errorf("featuresHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got.Features, tt.features) {
				t.Errorf("featuresHandler() = %v, want %v", got.Features, tt.features)
			}
		})
	}
}
