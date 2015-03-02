package base

import (
	"encoding/json"
	"testing"

	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/tests"
)

func TestSubscription(t *testing.T) {
	s := Subscription{}
	s.data.Link = "Link"

	tests.CheckString(t, "Subscription for Link", s.String())

	d := s.Data()

	tests.CheckString(t, "Link", d.Link)

	d = s.Data(data.Subscription{Link: "New link"})

	tests.CheckString(t, "New link", d.Link)

	tests.CheckBool(t, false, s.Validate() == nil)

	s.data.Link = ""
	tests.CheckBool(t, false, s.Validate() == nil)

	s.data.Link = "http://sugr.org"
	tests.CheckBool(t, false, s.Validate() == nil)

	s.data.FeedId = 1
	tests.CheckBool(t, true, s.Validate() == nil)

	ejson, eerr := json.Marshal(s.data)
	tests.CheckBool(t, true, eerr == nil)

	ajson, aerr := json.Marshal(s)
	tests.CheckBool(t, true, aerr == nil)

	tests.CheckBytes(t, ejson, ajson)
}
