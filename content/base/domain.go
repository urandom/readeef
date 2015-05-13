package base

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
)

type Domain struct {
	Error
	RepoRelated

	url *url.URL
}

var timeoutClient = readeef.NewTimeoutClient(5*time.Second, 5*time.Second)

func (d Domain) String() string {
	return d.url.Host
}

func (d *Domain) URL(u ...string) *url.URL {
	if d.HasErr() {
		return d.url
	}

	if len(u) > 0 {
		var err error

		if d.url, err = url.Parse(u[0]); err != nil {
			d.Err(err)
			return d.url
		}
	}

	return d.url
}

func (d Domain) Validate() error {
	if d.url == nil {
		return content.NewValidationError(errors.New("No url"))
	}

	return nil
}

func (d Domain) SupportsHTTPS() bool {
	if err := d.Validate(); err != nil {
		d.Err(err)
		return false
	}

	return d.url.Scheme == "https" || d.url.Host != "" && d.url.Scheme == ""
}

func (d Domain) CheckHTTPSSupport() bool {
	u := d.url
	if u.Scheme == "https" {
		return true
	}

	u.Scheme = "https"
	u.Path = ""
	u.RawQuery = ""

	if req, err := http.NewRequest("HEAD", u.String(), nil); err == nil {
		_, err := timeoutClient.Do(req)

		return err == nil
	}

	return false
}
