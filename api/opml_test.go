package api

import (
	"context"
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
									feedManager.EXPECT().AddFeedByLink(tt.discovered[i][j].Link).Return(tt.added[i][j], tt.addedErrs[i][j])
								}

								if tt.addedErrs[i][j] != nil {
									code = http.StatusInternalServerError
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
        <outline type="rss" text="text2" title="Item 2 title" xmlUrl="http://www.item2.com/rss" htmlUrl="http://www.item2.com"></outline>
    </body>
</opml>
`
)
