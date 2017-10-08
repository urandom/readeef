package repo_test

import (
	"reflect"
	"testing"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
)

func Test_extractRepo_Get(t *testing.T) {
	skipTest(t)
	setupArticle()

	tests := []struct {
		name      string
		article   content.Article
		want      content.Extract
		wantErr   bool
		noContent bool
	}{
		{"valid1", articles[0], content.Extract{ArticleID: articles[0].ID, Title: "title 1", Content: "content 1"}, false, false},
		{"valid2", articles[5], content.Extract{ArticleID: articles[5].ID, Title: "title 1", Content: "content 1", TopImage: "http://sugr.org/image", Language: "en"}, false, false},
		{"invalid", content.Article{}, content.Extract{}, true, false},
		{"no content", articles[7], content.Extract{}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := service.ExtractRepo()
			if !tt.noContent && !tt.wantErr {
				if err := r.Update(tt.want); err != nil {
					t.Errorf("extractRepo.Get() preliminary update error = %v", err)
					return
				}
			}

			got, err := r.Get(tt.article)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractRepo.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.noContent && errors.Cause(err) != content.ErrNoContent {
				t.Errorf("extractRepo.Get() error = %v, wanted no content", err)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractRepo.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractRepo_Update(t *testing.T) {
	skipTest(t)
	setupArticle()

	tests := []struct {
		name    string
		article content.Article
		extract content.Extract
		wantErr bool
	}{
		{"valid", articles[0], content.Extract{ArticleID: articles[0].ID, Title: "title 1", Content: "content 1"}, false},
		{"invalid", content.Article{}, content.Extract{Title: "title 2", Content: "content 2"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := service.ExtractRepo()
			if err := r.Update(tt.extract); (err != nil) != tt.wantErr {
				t.Errorf("extractRepo.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			got, err := r.Get(tt.article)
			if err != nil {
				t.Errorf("extractRepo.Update() post fetch error = %v", err)
				return
			}

			if !reflect.DeepEqual(got, tt.extract) {
				t.Errorf("extractRepo.Update() = %v, want %v", got, tt.extract)
			}
		})
	}
}
