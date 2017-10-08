package repo_test

import (
	"reflect"
	"testing"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
)

func Test_thumbnailRepo_Get(t *testing.T) {
	skipTest(t)
	setupArticle()

	tests := []struct {
		name      string
		article   content.Article
		want      content.Thumbnail
		wantErr   bool
		noContent bool
	}{
		{"valid1", articles[0], content.Thumbnail{ArticleID: articles[0].ID, Thumbnail: "thumb1"}, false, false},
		{"valid2", articles[5], content.Thumbnail{ArticleID: articles[5].ID, Thumbnail: "thumb2", Link: "http://sugr.org"}, false, false},
		{"invalid", content.Article{}, content.Thumbnail{}, true, false},
		{"no content", articles[7], content.Thumbnail{}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := service.ThumbnailRepo()
			if !tt.noContent && !tt.wantErr {
				if err := r.Update(tt.want); err != nil {
					t.Errorf("thumbnailRepo.Get() preliminary update error = %v", err)
					return
				}
			}
			got, err := r.Get(tt.article)
			if (err != nil) != tt.wantErr {
				t.Errorf("thumbnailRepo.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.noContent && errors.Cause(err) != content.ErrNoContent {
				t.Errorf("thumbnailRepo.Get() error = %v, wanted no content", err)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("thumbnailRepo.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_thumbnailRepo_Update(t *testing.T) {
	skipTest(t)
	setupArticle()

	tests := []struct {
		name      string
		article   content.Article
		thumbnail content.Thumbnail
		wantErr   bool
	}{
		{"valid", articles[0], content.Thumbnail{ArticleID: articles[0].ID, Thumbnail: "thumb1"}, false},
		{"invalid", content.Article{}, content.Thumbnail{Thumbnail: "thumb2", Link: "http://sugr.org"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := service.ThumbnailRepo()
			if err := r.Update(tt.thumbnail); (err != nil) != tt.wantErr {
				t.Errorf("thumbnailRepo.Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			got, err := r.Get(tt.article)
			if err != nil {
				t.Errorf("thumbnailRepo.Update() post fetch error = %v", err)
				return
			}

			if !reflect.DeepEqual(got, tt.thumbnail) {
				t.Errorf("thumbnailRepo.Update() = %v, want %v", got, tt.thumbnail)
			}
		})
	}
}
