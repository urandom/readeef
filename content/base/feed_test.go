package base

import (
	"encoding/json"
	"testing"

	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/parser"
	"github.com/urandom/readeef/tests"
)

func TestFeed(t *testing.T) {
	f := Feed{}
	f.data.Id = 1
	f.data.Title = "Title"

	tests.CheckString(t, "Title (1)", f.String())

	d := f.Data()

	tests.CheckString(t, "Title", d.Title)

	d = f.Data(data.Feed{Title: "New title", Description: "Desc"})

	tests.CheckString(t, "New title", d.Title)
	tests.CheckString(t, "Desc", d.Description)

	tests.CheckBool(t, false, f.Validate() == nil)

	f.data.Link = "foobar"
	tests.CheckBool(t, false, f.Validate() == nil)

	f.data.Link = "http://sugr.org"
	tests.CheckBool(t, true, f.Validate() == nil)

	tests.CheckInt64(t, 0, int64(len(f.ParsedArticles())))

	f.Refresh(parser.Feed{Title: "Title2", Articles: []parser.Article{parser.Article{Title: "Article title"}}})
	tests.CheckString(t, "Title2", f.data.Title)
	tests.CheckInt64(t, 1, int64(len(f.ParsedArticles())))

	tests.CheckString(t, "Article title", f.ParsedArticles()[0].Data().Title)

	ejson, eerr := json.Marshal(f.data)
	tests.CheckBool(t, true, eerr == nil)

	ajson, aerr := json.Marshal(f)
	tests.CheckBool(t, true, aerr == nil)

	tests.CheckBytes(t, ejson, ajson)
}
