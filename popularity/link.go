package popularity

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	dqLinkMatcher = regexp.MustCompile(`<a .*?href="(.*?)".*?>.*?</a>`)
	sqLinkMatcher = regexp.MustCompile(`<a .*?href='(.*?)'.*?>.*?</a>`)
)

func realLink(link, text string) string {
	u, err := url.Parse(link)
	if err != nil {
		return ""
	}
	u.RawQuery = ""
	link = u.String()

	if link, ok := processReddit(link, text); ok {
		return link
	}

	return link
}

func processReddit(link, text string) (string, bool) {
	if strings.Contains(link, "://reddit.com/r") || strings.Contains(link, "://www.reddit.com/r") {
		res := dqLinkMatcher.FindAllStringSubmatch(text, -1)
		for _, l := range res {
			if checkRedditLink(l[1], link) {
				return l[1], true
			}
		}

		res = sqLinkMatcher.FindAllStringSubmatch(text, -1)
		for _, l := range res {
			if checkRedditLink(l[1], link) {
				return l[1], true
			}
		}
	}

	return "", false
}

func checkRedditLink(link, original string) bool {
	if strings.Contains(link, "reddit.com/user") {
		return false
	}

	if strings.Contains(link, "reddit.com/") && strings.Contains(link, "/comments/") {
		return false
	}

	if link != original {
		return true
	}

	return false
}
