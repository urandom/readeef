package sql

import (
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/parser"
)

var (
	feed1 = content.Feed{ID: 1, Link: "http://sugr.org/1", Title: "feed 1"}
	feed2 = content.Feed{ID: 2, Link: "http://sugr.org/2", Title: "feed 2"}

	feedSync sync.Once
)

func Test_feedRepo_Get(t *testing.T) {
	skipTest(t)
	setupFeed()

	type args struct {
		id    content.FeedID
		login content.Login
	}
	tests := []struct {
		name    string
		args    args
		want    content.Feed
		wantErr bool
	}{
		{"get 1 without user", args{1, ""}, feed1, false},
		{"get 2 without user", args{2, ""}, feed2, false},
		{"get 1 with user1", args{1, "user1"}, feed1, false},
		{"get 2 with user1", args{2, "user1"}, feed2, false},
		{"get 2 with user2", args{2, "user2"}, feed2, false},
		{"get unknown without user", args{3, ""}, content.Feed{}, true},
		{"get unknown with user", args{3, "user1"}, content.Feed{}, true},
		{"get 1 with user2", args{1, "user2"}, content.Feed{}, true},
		{"get 2 with user3", args{2, "user3"}, content.Feed{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := service.FeedRepo()
			got, err := r.Get(tt.args.id, content.User{Login: tt.args.login})
			if (err != nil) != tt.wantErr {
				t.Errorf("feedRepo.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("feedRepo.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_feedRepo_FindByLink(t *testing.T) {
	skipTest(t)
	setupFeed()

	tests := []struct {
		name    string
		link    string
		want    content.Feed
		wantErr bool
	}{
		{"find 1", "http://sugr.org/1", feed1, false},
		{"find 2", "http://sugr.org/2", feed2, false},
		{"find non-existentg", "http://sugr.org/3", content.Feed{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := service.FeedRepo()
			got, err := r.FindByLink(tt.link)
			if (err != nil) != tt.wantErr {
				t.Errorf("feedRepo.FindByLink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("feedRepo.FindByLink() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_feedRepo_ForUser(t *testing.T) {
	skipTest(t)
	setupFeed()

	tests := []struct {
		name    string
		login   content.Login
		want    []content.Feed
		wantErr bool
	}{
		{"for user 1", "user1", []content.Feed{feed1, feed2}, false},
		{"for user 2", "user2", []content.Feed{feed2}, false},
		{"for user 3", "user3", []content.Feed{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := service.FeedRepo()
			got, err := r.ForUser(content.User{Login: tt.login})
			if (err != nil) != tt.wantErr {
				t.Errorf("feedRepo.ForUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) == 0 && len(tt.want) == 0 {
				return
			}

			sort.Slice(got, func(i, j int) bool {
				return got[i].ID < got[j].ID
			})

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("feedRepo.ForUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_feedRepo_ForTag(t *testing.T) {
	skipTest(t)
	setupFeed()

	type args struct {
		tag   content.Tag
		login content.Login
	}
	tests := []struct {
		name    string
		args    args
		want    []content.Feed
		wantErr bool
	}{
		{"for user 1 tag 1", args{tag1, "user1"}, []content.Feed{feed1}, false},
		{"for user 1 tag 2", args{tag2, "user1"}, []content.Feed{feed1, feed2}, false},
		{"for user 2 tag 1", args{tag1, "user2"}, []content.Feed{feed1}, false},
		{"for user 2 tag 2", args{tag2, "user2"}, nil, false},
		{"for user 3 tag 1", args{tag2, "user3"}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := service.FeedRepo()
			got, err := r.ForTag(tt.args.tag, content.User{Login: tt.args.login})
			if (err != nil) != tt.wantErr {
				t.Errorf("feedRepo.ForUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) == 0 && len(tt.want) == 0 {
				return
			}

			sort.Slice(got, func(i, j int) bool {
				return got[i].ID < got[j].ID
			})

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("feedRepo.ForUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_feedRepo_All(t *testing.T) {
	skipTest(t)
	setupFeed()

	r := service.FeedRepo()
	got, err := r.All()

	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 2 {
		t.Errorf("feedRepo.All() expected 2, got %v", len(got))
		return
	}

	expected := 2
	for _, f := range got {
		if f.ID == 1 || f.ID == 2 {
			expected--
		}
	}

	if expected != 0 {
		t.Errorf("feedRepo.All() expected both existing feeds, found %d", expected)
	}
}

func Test_feedRepo_IDs(t *testing.T) {
	skipTest(t)
	setupFeed()

	r := service.FeedRepo()
	got, err := r.IDs()

	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 2 {
		t.Errorf("feedRepo.IDs() expected 2, got %v", len(got))
		return
	}

	expected := 2
	for _, id := range got {
		if id == 1 || id == 2 {
			expected--
		}
	}

	if expected != 0 {
		t.Errorf("feedRepo.IDs() expected both existing feeds, found %d", expected)
	}
}

func Test_feedRepo_UpdateDelete(t *testing.T) {
	skipTest(t)
	setupFeed()

	tests := []struct {
		name    string
		feed    content.Feed
		parsed  parser.Feed
		want    []content.Article
		wantErr bool
	}{
		{"test1", content.Feed{Title: "title 3", Link: "http://sugr.org/3"}, parser.Feed{Title: "title 3"}, []content.Article{}, false},
		{"test2", content.Feed{Title: "title 3", Link: "http://sugr.org/3"}, parser.Feed{Title: "title 3", Articles: []parser.Article{
			{Title: "Article 100", Link: "http://sugr.org/3/article/100"},
			{Title: "Article 200", Link: "http://sugr.org/3/article/200"},
		}}, []content.Article{
			{Title: "Article 100", Link: "http://sugr.org/3/article/100"},
			{Title: "Article 200", Link: "http://sugr.org/3/article/200"},
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := service.FeedRepo()
			tt.feed.Refresh(tt.parsed)
			got, err := r.Update(&tt.feed)
			if (err != nil) != tt.wantErr {
				t.Errorf("feedRepo.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.feed.ID == 0 {
				t.Errorf("feedRepo.Update() no feed id")
				return
			}

			if len(got) == 0 && len(tt.want) == 0 {
				return
			}

			sort.Slice(got, func(i, j int) bool {
				return got[i].ID < got[j].ID
			})

			if len(got) != len(tt.want) {
				t.Errorf("feedRepo.Update() articles %d, want %d", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i].ID == 0 {
					t.Errorf("feedRepo.Update() article %v no id", got[i])
					return
				}

				if got[i].Title != tt.want[i].Title || got[i].Link != tt.want[i].Link {
					t.Errorf("feedRepo.Update() = %v, want %v", got[i], tt.want[i])
					return
				}
			}

			if err := r.Delete(tt.feed); err != nil {
				t.Errorf("feedRepo.Delete() error %v", err)
			}
		})
	}
}

func Test_feedRepo_Users(t *testing.T) {
	skipTest(t)
	setupFeed()

	tests := []struct {
		name   string
		feed   content.Feed
		attach []content.User
		detach int
	}{
		{"attach to user 1", content.Feed{Link: "http://sugr.org/10"}, []content.User{
			{Login: user1},
		}, 0},
		{"attach to user 1, and 2", content.Feed{Link: "http://sugr.org/11"}, []content.User{
			{Login: user1},
			{Login: user2},
		}, 0},
		{"attach to user 1, and 2 and detach from 1", content.Feed{Link: "http://sugr.org/11"}, []content.User{
			{Login: user1},
			{Login: user2},
		}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := service.FeedRepo()
			_, err := r.Update(&tt.feed)
			if err != nil {
				t.Errorf("feedRepo.Users() update feed error = %v", err)
				return
			}

			users, err := r.Users(tt.feed)
			if err != nil {
				t.Errorf("feedRepo.Users() error = %v", err)
				return
			}

			if len(users) > 0 {
				t.Errorf("feedRepo.Users() users len(%#v) != 0", users)
				return
			}

			if len(tt.attach) > 0 {
				for i, u := range tt.attach {
					if err := r.AttachTo(tt.feed, u); err != nil {
						t.Errorf("feedRepo.AttachTo() error = %v", err)
						return
					}

					if users, err := r.Users(tt.feed); err != nil {
						t.Errorf("feedRepo.Users() error = %v", err)
						return
					} else if len(users) != i+1 {
						t.Errorf("feedRepo.Users() count = %d, want %d", len(users), i+1)
						return
					}
				}

				for i := 0; i < tt.detach; i++ {
					if err := r.DetachFrom(tt.feed, tt.attach[i]); err != nil {
						t.Errorf("feedRepo.DetachFrom() error = %v", err)
						return
					}

					if users, err := r.Users(tt.feed); err != nil {
						t.Errorf("feedRepo.Users() error = %v", err)
						return
					} else if len(users) != len(tt.attach)-i-1 {
						t.Errorf("feedRepo.Users() count = %d, want %d", len(users), len(tt.attach)-i-1)
						return
					}
				}
			}

			if err := r.Delete(tt.feed); err != nil {
				t.Errorf("feedRepo.Delete() error %v", err)
			}
		})
	}
}

func Test_feedRepo_SetUserTags(t *testing.T) {
	skipTest(t)
	setupFeed()

	tests := []struct {
		name string
		feed content.Feed
		user content.Login
		tags []*content.Tag
	}{
		{"no tags", content.Feed{Link: "http://sugr.org/10"}, user1, []*content.Tag{}},
		{"simple", content.Feed{Link: "http://sugr.org/10"}, user1, []*content.Tag{
			{Value: "tag 10"},
			{Value: "tag 20"},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := service.FeedRepo()
			_, err := r.Update(&tt.feed)
			if err != nil {
				t.Errorf("feedRepo.SetUserTags() update feed error = %v", err)
				return
			}

			if err := r.SetUserTags(tt.feed, content.User{Login: tt.user}, tt.tags); err != nil {
				t.Errorf("feedRepo.SetUserTags() error = %v", err)
				return
			}

			tags, err := service.TagRepo().ForFeed(tt.feed, content.User{Login: tt.user})
			if err != nil {
				t.Errorf("tagRepo.ForFeed() error = %v", err)
				return
			}

			wanted := map[content.TagValue]struct{}{}
			for _, t := range tt.tags {
				wanted[t.Value] = struct{}{}
			}

			if len(tags) != len(tt.tags) {
				t.Errorf("tagRepo.ForFeed() len(tags) = %d, wanted %d", len(tags), len(tt.tags))
				return
			}

			for _, tag := range tags {
				if _, ok := wanted[tag.Value]; !ok {
					t.Errorf("tagRepo.ForFeed() tag %#v not found", tag)
					return
				}
			}

			if err := r.Delete(tt.feed); err != nil {
				t.Errorf("feedRepo.Delete() error %v", err)
			}
		})
	}
}
func createFeed(feed *content.Feed, users ...content.User) {
	r := service.FeedRepo()

	if _, err := r.Update(feed); err != nil {
		panic(err)
	}

	for _, u := range users {
		if err := r.AttachTo(*feed, u); err != nil {
			panic(err)
		}
	}
}

func setupFeed() {
	if skip {
		return
	}

	feedSync.Do(func() {
		setupUser()

		u1 := content.User{Login: user1}
		u2 := content.User{Login: user2}

		f1 := feed1
		f2 := feed2

		f1.Refresh(parser.Feed{Title: "feed 1", Articles: []parser.Article{
			{Title: "Article 1", Description: "Description 1", Link: "http://sugr.org/1/a/1"},
			{Title: "Article 2", Description: "Description 2", Link: "http://sugr.org/1/a/2"},
			{Title: "Article 3", Description: "Description 3", Link: "http://sugr.org/1/a/3"},
			{Title: "Article 4", Description: "Description 4", Link: "http://sugr.org/1/a/4"},
		}})

		f2.Refresh(parser.Feed{Title: "feed 2", Articles: []parser.Article{
			{Title: "Article 5", Description: "Description 5", Link: "http://sugr.org/2/a/5"},
			{Title: "Article 6", Description: "Description 6", Link: "http://sugr.org/2/a/6"},
			{Title: "Article 7", Description: "Description 7", Link: "http://sugr.org/2/a/7"},
			{Title: "Article 8", Description: "Description 8", Link: "http://sugr.org/2/a/8"},
			{Title: "Article 9", Description: "Description 9", Link: "http://sugr.org/2/a/9"},
		}})

		createFeed(&f1, u1)
		createFeed(&f2, u1, u2)

		if err := service.FeedRepo().SetUserTags(f1, u1, []*content.Tag{&tag1, &tag2}); err != nil {
			panic(err)
		}

		if err := service.FeedRepo().SetUserTags(f2, u1, []*content.Tag{&tag2}); err != nil {
			panic(err)
		}

		if err := service.FeedRepo().SetUserTags(f1, u2, []*content.Tag{&tag1}); err != nil {
			panic(err)
		}
	})
}
