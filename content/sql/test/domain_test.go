package test

import (
	"testing"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/tests"
)

type dom struct {
	content.Domain

	currentSupport bool
}

func (d dom) CheckHTTPSSupport() bool {
	return d.currentSupport
}

func TestDomain(t *testing.T) {
	realD := repo.Domain("http://sugr.org")

	d := &dom{Domain: realD}

	tests.CheckBool(t, false, d.HasErr(), d.Err())

	tests.CheckBool(t, false, d.SupportsHTTPS(), d.Err())

	d.currentSupport = true
	// Already checked, will return false
	tests.CheckBool(t, false, d.SupportsHTTPS(), d.Err())

	d.Domain = repo.Domain("http:/sugr.org")
	d.currentSupport = true
	tests.CheckBool(t, true, d.SupportsHTTPS(), d.Err())
}
