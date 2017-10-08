package repo_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/urandom/readeef/content"
)

var (
	tag1 = content.Tag{Value: "tag 1"}
	tag2 = content.Tag{Value: "tag 2"}
)

func Test_tagRepo_Get(t *testing.T) {
	skipTest(t)
	setupFeed()

	type args struct {
		id   content.TagID
		user content.Login
	}
	tests := []struct {
		name    string
		args    args
		want    content.Tag
		wantErr bool
	}{
		{"get tag 1 for user 1", args{tag1.ID, user1}, tag1, false},
		{"get tag 2 for user 1", args{tag2.ID, user1}, tag2, false},
		{"get tag 1 for user 2", args{tag1.ID, user2}, content.Tag{}, true},
		{"get tag 2 for user 2", args{tag2.ID, user2}, content.Tag{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := service.TagRepo()
			got, err := r.Get(tt.args.id, content.User{Login: tt.args.user})
			if (err != nil) != tt.wantErr {
				t.Errorf("tagRepo.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tagRepo.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tagRepo_ForUser(t *testing.T) {
	skipTest(t)
	setupFeed()

	tests := []struct {
		name    string
		user    content.Login
		want    []content.Tag
		wantErr bool
	}{
		{"get tags for user 1", user1, []content.Tag{tag1, tag2}, false},
		{"get tags for user 2", user2, []content.Tag{}, false},
		{"get tags for user 3", "user3", []content.Tag{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := service.TagRepo()
			got, err := r.ForUser(content.User{Login: tt.user})
			if (err != nil) != tt.wantErr {
				t.Errorf("tagRepo.ForUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			sort.Slice(got, func(i, j int) bool {
				return got[i].ID < got[j].ID
			})

			if len(got) == 0 && len(tt.want) == 0 {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tagRepo.ForUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tagRepo_ForFeed(t *testing.T) {
	skipTest(t)
	setupFeed()

	tests := []struct {
		name    string
		feed    content.Feed
		user    content.Login
		want    []content.Tag
		wantErr bool
	}{
		{"get tags for feed 1 user 1", feed1, user1, []content.Tag{tag1, tag2}, false},
		{"get tags for feed 1 user 2", feed1, user2, []content.Tag{}, false},
		{"get tags for feed 1 user 3", feed1, "user3", []content.Tag{}, false},
		{"get tags for feed 2 user 1", feed2, user1, []content.Tag{tag2}, false},
		{"get tags for feed 2 user 2", feed2, user2, []content.Tag{}, false},
		{"get tags for feed 2 user 3", feed2, "user3", []content.Tag{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := service.TagRepo()
			got, err := r.ForFeed(tt.feed, content.User{Login: tt.user})
			if (err != nil) != tt.wantErr {
				t.Errorf("tagRepo.ForFeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			sort.Slice(got, func(i, j int) bool {
				return got[i].ID < got[j].ID
			})

			if len(got) == 0 && len(tt.want) == 0 {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tagRepo.ForFeed() = %v, want %v", got, tt.want)
			}
		})
	}
}
