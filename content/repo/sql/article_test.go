package sql

import (
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/urandom/readeef/content"
)

var (
	articles    []content.Article
	articleSync sync.Once
)

func setupArticle() {
	if skip {
		return
	}

	articleSync.Do(func() {
		setupFeed()

		r := service.ArticleRepo()

		var err error
		articles, err = r.All()
		if err != nil {
			panic(err)
		}

		sort.Slice(articles, func(i, j int) bool {
			return strings.Compare(articles[i].Title, articles[j].Title) == -1
		})

		unreadU1 := []content.ArticleID{}
		unreadU2 := []content.ArticleID{}
		favorU1 := []content.ArticleID{}
		favorU2 := []content.ArticleID{}

		for _, a := range articles {
			switch a.Title {
			case "Article 1", "Article 2":
				unreadU1 = append(unreadU1, a.ID)
				favorU1 = append(favorU1, a.ID)
			case "Article 4":
				unreadU1 = append(unreadU1, a.ID)
			case "Article 5", "Article 6":
				unreadU1 = append(unreadU1, a.ID)
				unreadU2 = append(unreadU2, a.ID)
			case "Article 7":
				unreadU2 = append(unreadU2, a.ID)
				favorU2 = append(favorU2, a.ID)
			case "Article 8":
				unreadU1 = append(unreadU1, a.ID)
				favorU1 = append(favorU1, a.ID)
			case "Article 9":
				unreadU2 = append(unreadU2, a.ID)
			}
		}

		if err = r.Read(false, content.User{Login: user1}, content.IDs(unreadU1)); err != nil {
			panic(err)
		}
		if err = r.Read(false, content.User{Login: user2}, content.IDs(unreadU2)); err != nil {
			panic(err)
		}
		if err = r.Favor(true, content.User{Login: user1}, content.IDs(favorU1)); err != nil {
			panic(err)
		}
		if err = r.Favor(true, content.User{Login: user2}, content.IDs(favorU2)); err != nil {
			panic(err)
		}
	})
}

func Test_articleRepo_ForUser(t *testing.T) {
	skipTest(t)
	setupArticle()

	type args struct {
		user content.Login
		opts []content.QueryOpt
	}
	tests := []struct {
		name    string
		args    args
		want    []content.ArticleID
		wantErr bool
	}{
		{"all user1 articles", args{user: user1}, []content.ArticleID{
			articles[0].ID, articles[1].ID, articles[2].ID, articles[3].ID, articles[4].ID,
			articles[5].ID, articles[6].ID, articles[7].ID, articles[8].ID,
		}, false},
		{"all user2 articles", args{user: user2}, []content.ArticleID{
			articles[4].ID, articles[5].ID, articles[6].ID, articles[7].ID, articles[8].ID,
		}, false},
		{"all user3 articles", args{user: "user3"}, []content.ArticleID{}, false},
		{"all empty user articles", args{user: ""}, []content.ArticleID{}, true},
		{"read user1 articles", args{user1, []content.QueryOpt{content.ReadOnly}}, []content.ArticleID{
			articles[2].ID, articles[6].ID, articles[8].ID,
		}, false},
		{"unread user1 articles", args{user1, []content.QueryOpt{content.UnreadOnly}}, []content.ArticleID{
			articles[0].ID, articles[1].ID, articles[3].ID, articles[4].ID,
			articles[5].ID, articles[7].ID,
		}, false},
		{"read user1 articles from feed 2", args{user1, []content.QueryOpt{content.ReadOnly, content.FeedIDs([]content.FeedID{feed2.ID})}}, []content.ArticleID{
			articles[6].ID, articles[8].ID,
		}, false},
		{"specific ids for user 2", args{user2, []content.QueryOpt{content.IDs([]content.ArticleID{articles[4].ID, articles[7].ID})}}, []content.ArticleID{
			articles[4].ID, articles[7].ID,
		}, false},
		{"id range with paging", args{user1, []content.QueryOpt{content.IDRange(articles[2].ID, articles[7].ID), content.Paging(2, 1)}}, []content.ArticleID{
			articles[4].ID, articles[5].ID,
		}, false},
		{"specific feed id for user 1", args{user1, []content.QueryOpt{content.FeedIDs([]content.FeedID{feed1.ID})}}, []content.ArticleID{
			articles[0].ID, articles[1].ID, articles[2].ID, articles[3].ID,
		}, false},
		{"time range for user 2", args{user2, []content.QueryOpt{content.TimeRange(time.Now().Add(-5*time.Hour), time.Now().Add(-2*time.Hour))}}, []content.ArticleID{
			articles[5].ID, articles[6].ID, articles[7].ID,
		}, false},
		{"older and unread first for user 2", args{user2, []content.QueryOpt{content.UnreadFirst, content.Sorting(content.SortByDate, content.DescendingOrder)}}, []content.ArticleID{
			articles[8].ID, articles[6].ID, articles[5].ID, articles[4].ID, articles[7].ID,
		}, false},
		{"favorite for user 1", args{user1, []content.QueryOpt{content.FavoriteOnly}}, []content.ArticleID{
			articles[0].ID, articles[1].ID, articles[7].ID,
		}, false},
		{"untagged for user 1", args{user1, []content.QueryOpt{content.UntaggedOnly}}, []content.ArticleID{}, false},
		{"untagged for user 2", args{user2, []content.QueryOpt{content.UntaggedOnly}}, []content.ArticleID{
			articles[4].ID, articles[5].ID, articles[6].ID, articles[7].ID, articles[8].ID,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := service.ArticleRepo()
			got, err := r.ForUser(content.User{Login: tt.args.user}, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("articleRepo.ForUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) == 0 && len(tt.want) == 0 {
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("articleRepo.ForUser() article length mismatch = %d, want = %d", len(got), len(tt.want))
				return
			}

			idSet := map[content.ArticleID]struct{}{}
			for _, id := range tt.want {
				idSet[id] = struct{}{}
			}

			for _, a := range got {
				if _, ok := idSet[a.ID]; !ok {
					t.Errorf("articleRepo.ForUser() unknown article = %#v", a)
					return
				}
			}
		})
	}
}
