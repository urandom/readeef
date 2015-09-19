package base

import (
	"encoding/json"
	"testing"

	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/tests"
)

func TestTag(t *testing.T) {
	tag := Tag{}
	tag.data.Value = "Link"

	tests.CheckString(t, "Link", tag.String())

	d := tag.Data()

	tests.CheckString(t, "Link", string(d.Value))

	d = tag.Data(data.Tag{Value: ""})

	tests.CheckString(t, "", string(d.Value))

	tests.CheckBool(t, false, tag.Validate() == nil)

	tag.data.Value = "tag"
	tests.CheckBool(t, false, tag.Validate() == nil)

	ejson, eerr := json.Marshal(tag.data.Value)
	tests.CheckBool(t, true, eerr == nil)

	ajson, aerr := json.Marshal(tag)
	tests.CheckBool(t, true, aerr == nil)

	tests.CheckBytes(t, ejson, ajson)
}
