package base

import (
	"encoding/json"
	"testing"

	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/tests"
)

func TestArticle(t *testing.T) {
	a := Article{}
	a.data.Title = "Title"
	a.data.Id = data.ArticleId(1)

	tests.CheckString(t, "Title (1)", a.String())

	d := a.Data()

	tests.CheckString(t, "Title", d.Title)

	d = a.Data(data.Article{Title: "New title", Description: "Desc"})

	tests.CheckString(t, "New title", d.Title)
	tests.CheckString(t, "Desc", d.Description)

	tests.CheckBool(t, false, a.Validate() == nil)

	d.Link = "http://sugr.org/en/"
	d.FeedId = 42
	a.Data(d)

	tests.CheckBool(t, true, a.Validate() == nil)

	ejson, eerr := json.Marshal(d)
	tests.CheckBool(t, true, eerr == nil)

	ajson, aerr := json.Marshal(a)
	tests.CheckBool(t, true, aerr == nil)

	tests.CheckBytes(t, ejson, ajson)
}
