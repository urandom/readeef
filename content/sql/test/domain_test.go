package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/urandom/readeef/tests"
)

func TestDomain(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer ts.Close()

	d := repo.Domain(ts.URL)

	tests.CheckBool(t, false, d.HasErr(), d.Err())

	tests.CheckBool(t, false, d.SupportsHTTPS(), d.Err())
}
