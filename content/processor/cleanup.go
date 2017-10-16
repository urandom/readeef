package processor

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/urandom/readeef/log"
	"github.com/urandom/readeef/parser"
	"github.com/urandom/readeef/pool"
	"golang.org/x/net/html"
)

type Cleanup struct {
	log log.Log
}

func NewCleanup(l log.Log) Cleanup {
	return Cleanup{log: l}
}

var (
	iframeFixer = regexp.MustCompile(`(<iframe [^>]+)/>`)
)

func (p Cleanup) ProcessFeed(f parser.Feed) parser.Feed {
	p.log.Infof("Cleaning up feed '%s'\n", f.Title)

	for i := range f.Articles {
		f.Articles[i].Description = strings.TrimSpace(f.Articles[i].Description)

		// html.Parse breaks on self-closing iframe tags
		f.Articles[i].Description =
			iframeFixer.ReplaceAllString(f.Articles[i].Description, "$1></iframe>")

		if nodes, err := html.ParseFragment(strings.NewReader(f.Articles[i].Description), nil); err == nil {
			if nodesCleanup(nodes) {
				if len(nodes) == 0 {
					break
				}

				buf := pool.Buffer.Get()
				defer pool.Buffer.Put(buf)

				for _, n := range nodes {
					err = html.Render(buf, n)
					if err != nil {
						break
					}
				}

				content := buf.String()

				// net/http tries to provide valid html, adding html, head and body tags
				content = content[strings.Index(content, "<body>")+6 : strings.LastIndex(content, "</body>")]
				f.Articles[i].Description = content
			}
		}
	}

	return f
}

func nodesCleanup(nodes []*html.Node) bool {
	changed := false

	for _, n := range nodes {
		if n.Type == html.ElementNode && n.Data == "script" || n.Type == html.CommentNode {
			if n.Parent != nil {
				n.Parent.RemoveChild(n)
				changed = true
			}
			break
		}

		if n.Type == html.ElementNode {
			// Remove all 'on*' attributes, xmlns attributes, and any that contain 'javascript:'
			attrs := []html.Attribute{}
			for _, a := range n.Attr {
				if strings.HasPrefix(a.Key, "on") || a.Key == "xmlns" {
					changed = true
					break
				}

				i := strings.Index(a.Val, "javascript:")
				if i != -1 {
					onlySpace := true
					if i > 0 {
						for _, r := range a.Val[:i] {
							if !unicode.IsSpace(r) {
								onlySpace = false
								break
							}
						}
					}

					if onlySpace {
						changed = true
						break
					}
				}

				attrs = append(attrs, a)
			}

			n.Attr = attrs

			if n.Data == "iframe" {
			}

			// Add a target attribute to the article links
			if n.Data == "a" {
				var attr *html.Attribute

				for i, a := range n.Attr {
					if a.Key == "target" {
						attr = &n.Attr[i]
						break
					}
				}

				val := "feed-article"
				if attr == nil {
					n.Attr = append(n.Attr, html.Attribute{Key: "target", Val: val})
				} else {
					attr.Val = val
				}
				changed = true
			}
		}

		children := []*html.Node{}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			children = append(children, c)
		}

		if len(children) > 0 {
			if nodesCleanup(children) {
				changed = true
			}
		}
	}

	return changed
}
