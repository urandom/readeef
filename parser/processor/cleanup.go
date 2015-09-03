package processor

import (
	"strings"
	"unicode"

	"github.com/urandom/readeef/parser"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/util"
	"golang.org/x/net/html"
)

type Cleanup struct {
	logger webfw.Logger
}

func NewCleanup(l webfw.Logger) Cleanup {
	return Cleanup{logger: l}
}

func (p Cleanup) Process(f parser.Feed) parser.Feed {
	p.logger.Infof("Cleaning up feed '%s'\n", f.Title)

	for i := range f.Articles {
		f.Articles[i].Description = strings.TrimSpace(f.Articles[i].Description)

		if nodes, err := html.ParseFragment(strings.NewReader(f.Articles[i].Description), nil); err == nil {
			if nodesCleanup(nodes) {
				if len(nodes) == 0 {
					break
				}

				buf := util.BufferPool.GetBuffer()
				defer util.BufferPool.Put(buf)

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
			// Remove all 'on*' attributes, and any that contain 'javascript:'
			attrs := []html.Attribute{}
			for _, a := range n.Attr {
				if strings.HasPrefix(a.Key, "on") {
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
