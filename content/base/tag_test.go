package base

import (
	"encoding/json"
	"testing"

	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/tests"
)

func TestTag(t *testing.T) {
	tag := Tag{}
	tag.value = "Link"

	tests.CheckString(t, "Link", tag.String())

	d := tag.Value()

	tests.CheckString(t, "Link", string(d))

	d = tag.Value(data.TagValue(""))

	tests.CheckString(t, "", string(d))

	tests.CheckBool(t, false, tag.Validate() == nil)

	tag.value = "tag"
	tests.CheckBool(t, false, tag.Validate() == nil)

	ejson, eerr := json.Marshal(tag.value)
	tests.CheckBool(t, true, eerr == nil)

	ajson, aerr := json.Marshal(tag)
	tests.CheckBool(t, true, aerr == nil)

	tests.CheckBytes(t, ejson, ajson)
}
