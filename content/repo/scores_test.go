package repo_test

import (
	"reflect"
	"testing"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
)

func Test_scoresRepo_Get(t *testing.T) {
	skipTest(t)
	setupArticle()

	tests := []struct {
		name      string
		article   content.Article
		want      content.Scores
		wantErr   bool
		noContent bool
	}{
		{"valid1", articles[0], content.Scores{ArticleID: articles[0].ID, Score: 11}, false, false},
		{"valid2", articles[5], content.Scores{ArticleID: articles[5].ID, Score: 5, Score1: 4, Score2: 2, Score3: 1}, false, false},
		{"invalid", content.Article{}, content.Scores{}, true, false},
		{"no content", articles[7], content.Scores{}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := service.ScoresRepo()
			if !tt.noContent && !tt.wantErr {
				if err := r.Update(tt.want); err != nil {
					t.Errorf("scoresRepo.Get() preliminary update error = %v", err)
					return
				}
			}
			got, err := r.Get(tt.article)
			if (err != nil) != tt.wantErr {
				t.Errorf("scoresRepo.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.noContent && errors.Cause(err) != content.ErrNoContent {
				t.Errorf("scoresRepo.Get() error = %v, wanted no content", err)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("scoresRepo.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scoresRepo_Update(t *testing.T) {
	skipTest(t)
	setupArticle()

	tests := []struct {
		name    string
		article content.Article
		scores  content.Scores
		wantErr bool
	}{
		{"valid", articles[0], content.Scores{ArticleID: articles[0].ID, Score: 15}, false},
		{"invalid", content.Article{}, content.Scores{Score: 8, Score1: 5}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := service.ScoresRepo()
			if err := r.Update(tt.scores); (err != nil) != tt.wantErr {
				t.Errorf("scoresRepo.Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			got, err := r.Get(tt.article)
			if err != nil {
				t.Errorf("scoresRepo.Update() post fetch error = %v", err)
				return
			}

			if !reflect.DeepEqual(got, tt.scores) {
				t.Errorf("scoresRepo.Update() = %v, want %v", got, tt.scores)
			}
		})
	}
}
