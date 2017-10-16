package processor

import (
	"strings"
	"testing"

	"github.com/urandom/readeef/parser"
)

func TestCleanup_ProcessFeed(t *testing.T) {
	tests := []struct {
		name  string
		input parser.Feed
		want  parser.Feed
	}{
		{"data1", parser.Feed{Articles: []parser.Article{
			{Description: data1},
		}}, parser.Feed{Articles: []parser.Article{
			{Description: exp1},
		}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Cleanup{log: logger}

			got := p.ProcessFeed(tt.input)

			if len(got.Articles) != len(tt.want.Articles) {
				t.Errorf("Cleanup.ProcessFeed() len() = %d, want %d", len(got.Articles), len(tt.want.Articles))
				return
			}

			for i := range got.Articles {
				g := strings.TrimSpace(got.Articles[i].Description)
				w := strings.TrimSpace(tt.want.Articles[i].Description)

				if g != w {
					t.Errorf("Cleanup.ProcessFeed() = \n%s, want \n%s", g, w)
					return
				}
			}
		})
	}
}

const (
	data1 = `
	<div xmlns="http://www.w3.org/1999/xhtml"><p>If you’re running Wayland and have a touchpad capable of multi-touch, Builder (Nightly) now lets you do fun stuff like the following video demonstrates.</p>
<p>Just three-finger-swipe left or right to move the document. Content is sticky-to-fingers, which is my expectation when using gestures.</p>
<p>It might also work on a touchscreen, but I haven’t tried.</p>
<p><iframe allowfullscreen="allowfullscreen" frameborder="0" height="495" src="https://www.youtube.com/embed/gRRfjcGFcbw?feature=oembed" width="660"/></p></div>
	`
	exp1 = `
	<div><p>If you’re running Wayland and have a touchpad capable of multi-touch, Builder (Nightly) now lets you do fun stuff like the following video demonstrates.</p>
<p>Just three-finger-swipe left or right to move the document. Content is sticky-to-fingers, which is my expectation when using gestures.</p>
<p>It might also work on a touchscreen, but I haven’t tried.</p>
<p><iframe allowfullscreen="allowfullscreen" frameborder="0" height="495" src="https://www.youtube.com/embed/gRRfjcGFcbw?feature=oembed" width="660"></iframe></p></div>
	`
)
