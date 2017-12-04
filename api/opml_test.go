package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/mock_repo"
	"github.com/urandom/readeef/parser"
)

func Test_importOPML(t *testing.T) {
	tests := []struct {
		name            string
		hasUser         bool
		form            url.Values
		numInput        int
		hasParseErr     bool
		feeds           []content.Feed
		feedsErr        error
		findErrs        []error
		discovered      [][]content.Feed
		discoveredErrs  []error
		addedCalls      [][]bool
		added           [][]content.Feed
		addedErrs       [][]error
		attachErrs      [][]error
		setUserTagsErrs [][]error
	}{
		{"no user", false, nil, 0, false, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil},
		{"parse error", true, url.Values{"opml": []string{"not-xml"}}, 0, true, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil},
		{"feeds err", true, url.Values{"opml": []string{singleOmplXML}}, 1, false, nil, errors.New("feeds err"), nil, nil, nil, nil, nil, nil, nil, nil},
		{"exists", true, url.Values{"opml": []string{singleOmplXML}}, 1, false, nil, nil, []error{nil}, nil, nil, nil, nil, nil, nil, nil},
		{
			name:            "discovered errs",
			hasUser:         true,
			form:            url.Values{"opml": []string{singleOmplXML}},
			numInput:        1,
			hasParseErr:     false,
			feeds:           []content.Feed{{ID: 1}},
			feedsErr:        nil,
			findErrs:        []error{content.ErrNoContent},
			discovered:      [][]content.Feed{nil},
			discoveredErrs:  []error{errors.New("discovered err")},
			addedCalls:      nil,
			added:           nil,
			addedErrs:       nil,
			attachErrs:      nil,
			setUserTagsErrs: nil,
		},
		{
			name:            "bad link",
			hasUser:         true,
			form:            url.Values{"opml": []string{singleOmplXML}},
			numInput:        1,
			hasParseErr:     false,
			feeds:           []content.Feed{{ID: 1}},
			feedsErr:        nil,
			findErrs:        []error{content.ErrNoContent},
			discovered:      [][]content.Feed{{{Link: "://example.com"}}},
			discoveredErrs:  []error{nil},
			addedCalls:      [][]bool{{false}},
			added:           [][]content.Feed{{{}}},
			addedErrs:       [][]error{{errors.New("bad link err")}},
			attachErrs:      nil,
			setUserTagsErrs: nil,
		},
		{
			name:            "relative link",
			hasUser:         true,
			form:            url.Values{"opml": []string{singleOmplXML}},
			numInput:        1,
			hasParseErr:     false,
			feeds:           []content.Feed{{ID: 1}},
			feedsErr:        nil,
			findErrs:        []error{content.ErrNoContent},
			discovered:      [][]content.Feed{{{Link: "example.com"}}},
			discoveredErrs:  []error{nil},
			addedCalls:      [][]bool{{false}},
			added:           [][]content.Feed{{{}}},
			addedErrs:       [][]error{{errors.New("relative link err")}},
			attachErrs:      nil,
			setUserTagsErrs: nil,
		},
		{
			name:            "manager add err",
			hasUser:         true,
			form:            url.Values{"opml": []string{singleOmplXML}},
			numInput:        1,
			hasParseErr:     false,
			feeds:           []content.Feed{{ID: 1}},
			feedsErr:        nil,
			findErrs:        []error{content.ErrNoContent},
			discovered:      [][]content.Feed{{{Link: "https://example.com"}}},
			discoveredErrs:  []error{nil},
			addedCalls:      [][]bool{{true}},
			added:           [][]content.Feed{{{}}},
			addedErrs:       [][]error{{errors.New("add err")}},
			attachErrs:      nil,
			setUserTagsErrs: nil,
		},
		{
			name:            "error attaching",
			hasUser:         true,
			form:            url.Values{"opml": []string{singleOmplXML}},
			numInput:        1,
			hasParseErr:     false,
			feeds:           []content.Feed{{ID: 1}},
			feedsErr:        nil,
			findErrs:        []error{content.ErrNoContent},
			discovered:      [][]content.Feed{{{Link: "https://example.com"}, {Link: "https://example2.com"}}},
			discoveredErrs:  []error{nil},
			addedCalls:      [][]bool{{true, true}},
			added:           [][]content.Feed{{{Link: "https://example.com"}, {Link: "https://example2.com"}}},
			addedErrs:       [][]error{{nil, nil}},
			attachErrs:      [][]error{{errors.New("attach err"), nil}},
			setUserTagsErrs: nil,
		},
		{
			name:            "successful addition",
			hasUser:         true,
			form:            url.Values{"opml": []string{singleOmplXML}},
			numInput:        1,
			hasParseErr:     false,
			feeds:           []content.Feed{{ID: 1}},
			feedsErr:        nil,
			findErrs:        []error{content.ErrNoContent},
			discovered:      [][]content.Feed{{{Link: "https://example.com"}, {Link: "https://example2.com"}}},
			discoveredErrs:  []error{nil},
			addedCalls:      [][]bool{{true, true}},
			added:           [][]content.Feed{{{Link: "https://example.com"}, {Link: "https://example2.com"}}},
			addedErrs:       [][]error{{nil, nil}},
			attachErrs:      [][]error{{nil, nil}},
			setUserTagsErrs: nil,
		},
		{
			name:            "feeds with tags",
			hasUser:         true,
			form:            url.Values{"opml": []string{twoOmplXML}},
			numInput:        2,
			hasParseErr:     false,
			feeds:           []content.Feed{{ID: 1}},
			feedsErr:        nil,
			findErrs:        []error{content.ErrNoContent, content.ErrNoContent},
			discovered:      [][]content.Feed{{{Link: "https://example.com"}, {Link: "https://example2.com"}}, {{Link: "https://example3.com"}}},
			discoveredErrs:  []error{nil, nil},
			addedCalls:      [][]bool{{true, true}, {true}},
			added:           [][]content.Feed{{{Link: "https://example.com"}, {Link: "https://example2.com"}}, {{Link: "https://example3.com"}}},
			addedErrs:       [][]error{{nil, nil}, {nil}},
			attachErrs:      [][]error{{nil, nil}, {nil}},
			setUserTagsErrs: [][]error{{nil, nil}, {nil}},
		},
		{
			name:            "set tag err",
			hasUser:         true,
			form:            url.Values{"opml": []string{twoOmplXML}},
			numInput:        2,
			hasParseErr:     false,
			feeds:           []content.Feed{{ID: 1}},
			feedsErr:        nil,
			findErrs:        []error{content.ErrNoContent, content.ErrNoContent},
			discovered:      [][]content.Feed{{{Link: "https://example.com"}, {Link: "https://example2.com"}}, {{Link: "https://example3.com"}}},
			discoveredErrs:  []error{nil, nil},
			addedCalls:      [][]bool{{true, true}, {true}},
			added:           [][]content.Feed{{{Link: "https://example.com"}, {Link: "https://example2.com"}}, {{Link: "https://example3.com"}}},
			addedErrs:       [][]error{{nil, nil}, {nil}},
			attachErrs:      [][]error{{nil, nil}, {nil}},
			setUserTagsErrs: [][]error{{nil, nil}, {errors.New("set tag err")}},
		},
		{
			name:            "dry run",
			hasUser:         true,
			form:            url.Values{"opml": []string{singleOmplXML}, "dryRun": []string{""}},
			numInput:        1,
			hasParseErr:     false,
			feeds:           []content.Feed{{ID: 1}},
			feedsErr:        nil,
			findErrs:        []error{content.ErrNoContent},
			discovered:      [][]content.Feed{{{Link: "https://example.com"}}},
			discoveredErrs:  []error{nil},
			addedCalls:      [][]bool{{false}},
			added:           [][]content.Feed{{{Link: "https://example.com"}}},
			addedErrs:       [][]error{{nil}},
			attachErrs:      [][]error{{}},
			setUserTagsErrs: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			feedRepo := mock_repo.NewMockFeed(ctrl)
			feedManager := NewMockfeedManager(ctrl)

			r := httptest.NewRequest("POST", "/", strings.NewReader(tt.form.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			r.ParseForm()
			w := httptest.NewRecorder()

			code := http.StatusBadRequest
			total := tt.numInput
			if tt.hasUser {
				user := content.User{Login: "test"}
				r = r.WithContext(context.WithValue(r.Context(), userKey, user))

				if !tt.hasParseErr {
					feedRepo.EXPECT().ForUser(userMatcher{user}).Return(tt.feeds, tt.feedsErr)

					if tt.feedsErr != nil {
						code = http.StatusInternalServerError
					} else {
						code = http.StatusOK
						for i, err := range tt.findErrs {
							if err == nil {
								total--
							}

							link := fmt.Sprintf("http://www.item%d.com/rss", i+1)
							feedRepo.EXPECT().FindByLink(link).Return(content.Feed{}, err)
						}

						if total > 0 {
							for i, err := range tt.discoveredErrs {
								link := fmt.Sprintf("http://www.item%d.com/rss", i+1)
								if err != nil {
									total--
								}

								feedManager.EXPECT().DiscoverFeeds(link).Return(tt.discovered[i], err)
							}
						}

						for i := range tt.addedErrs {
							if tt.discoveredErrs[i] != nil {
								continue
							}

							for j := range tt.addedErrs[i] {
								if tt.addedCalls[i][j] {
									link := tt.discovered[i][j].Link
									if link == "https://example3.com" {
										link = "https://example3.com#cat1"
										feedRepo.EXPECT().SetUserTags(tt.added[i][j], userMatcher{user}, []*content.Tag{&content.Tag{Value: "cat1"}}).Return(tt.setUserTagsErrs[i][j])

										if tt.setUserTagsErrs[i][j] != nil {
											code = http.StatusInternalServerError
										}
									}
									feedManager.EXPECT().AddFeedByLink(link).Return(tt.added[i][j], tt.addedErrs[i][j])
								}

								if tt.addedErrs[i][j] == nil {
									if len(tt.attachErrs[i]) > j {
										feedRepo.EXPECT().AttachTo(tt.added[i][j], userMatcher{user}).Return(tt.attachErrs[i][j])

										if tt.attachErrs[i][j] != nil {
											code = http.StatusInternalServerError
											break
										}
									}
								} else {
									code = http.StatusInternalServerError
									break
								}
							}
						}
					}
				}
			}

			importOPML(feedRepo, feedManager, logger).ServeHTTP(w, r)

			if w.Code != code {
				t.Errorf("importOPML() code = %v, want %v", w.Code, code)
				return
			}
		})
	}
}

const (
	singleOmplXML = `
<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.1">
    <head>
        <title>
			OPML title
		</title>
    </head>
    <body>
        <outline type="rss" text="Item 1 text" title="Item 1 title" xmlUrl="http://www.item1.com/rss" htmlUrl="http://www.item1.com"></outline>
    </body>
</opml>
`
	twoOmplXML = `
<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.1">
    <head>
        <title>
			OPML title
		</title>
    </head>
    <body>
        <outline type="rss" text="text1" title="Item 1 title" xmlUrl="http://www.item1.com/rss" htmlUrl="http://www.item1.com"></outline>
        <outline type="rss" text="text2" title="Item 2 title" xmlUrl="http://www.item2.com/rss" htmlUrl="http://www.item2.com" category="cat1"></outline>
    </body>
</opml>
`
)

func Test_exportOPML(t *testing.T) {
	tests := []struct {
		name         string
		hasUser      bool
		userFeeds    []content.Feed
		userFeedsErr error
		tags         [][]content.Tag
		tagsErr      []error
	}{
		{name: "no user", hasUser: false},
		{name: "user feeds err", hasUser: true, userFeedsErr: errors.New("user feeds err")},
		{
			name:      "feeds without tags",
			hasUser:   true,
			userFeeds: []content.Feed{{Link: "http://example.com"}, {Link: "http://example2.com"}},
			tags:      [][]content.Tag{{}, {}},
			tagsErr:   []error{nil, nil},
		},
		{
			name:      "tag error",
			hasUser:   true,
			userFeeds: []content.Feed{{Link: "http://example.com"}, {Link: "http://example2.com"}},
			tags:      [][]content.Tag{{{Value: "tag1"}, {Value: "tag2"}}, {}},
			tagsErr:   []error{nil, errors.New("tag err")},
		},
	}

	type data struct {
		Opml string `json:"opml"`
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			service := mock_repo.NewMockService(ctrl)
			feedRepo := mock_repo.NewMockFeed(ctrl)
			tagRepo := mock_repo.NewMockTag(ctrl)

			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			code := http.StatusOK
			if tt.hasUser {
				user := content.User{Login: "test"}
				r = r.WithContext(context.WithValue(r.Context(), userKey, user))

				service.EXPECT().FeedRepo().Return(feedRepo)
				feedRepo.EXPECT().ForUser(userMatcher{user}).Return(tt.userFeeds, tt.userFeedsErr)

				if tt.userFeedsErr == nil {
					service.EXPECT().TagRepo().Return(tagRepo)
					for i, f := range tt.userFeeds {
						tagRepo.EXPECT().ForFeed(f, userMatcher{user}).Return(tt.tags[i], tt.tagsErr[i])
						if tt.tagsErr[i] != nil {
							code = http.StatusInternalServerError
							break
						}
					}
				} else {
					code = http.StatusInternalServerError
				}
			} else {
				code = http.StatusBadRequest
			}

			exportOPML(service, logger).ServeHTTP(w, r)

			if w.Code != code {
				t.Errorf("exportOPML() code = %v, want %v", w.Code, code)
				return
			}

			var got data
			err := json.Unmarshal(w.Body.Bytes(), &got)
			if (err != nil) && (code == http.StatusOK) {
				t.Errorf("exportOPML() response parse error = %v", err)
				return
			}

			if err == nil {
				opml, err := parser.ParseOpml([]byte(got.Opml))
				if err != nil {
					t.Errorf("exportOPML() opml parse error = %v", err)
					return
				}

				if len(opml.Feeds) != len(tt.userFeeds) {
					t.Errorf("exportOPML() opml.Feeds = %v, want = %v", opml.Feeds, tt.userFeeds)
					return
				}

				for i := range opml.Feeds {
					if opml.Feeds[i].URL != tt.userFeeds[i].Link {
						t.Errorf("exportOPML() opml.Feed = %v, want = %v", opml.Feeds[i], tt.userFeeds[i])
						return
					}
				}
			}

		})
	}
}
