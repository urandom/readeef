package parser

import (
	"encoding/xml"
	"net/http"
	"strings"
	"time"
)

func ParseFeed(b []byte) (feed, error) {
	f, err := ParseRss2(b)

	if err != nil {
		if _, ok := err.(xml.UnmarshalError); ok {
			f, err = ParseAtom(b)
			if err != nil {
				if _, ok := err.(xml.UnmarshalError); ok {
					f, err = ParseRss1(b)
					if err != nil {
						return feed{}, err
					}
				} else {
					return f, err
				}
			}
		} else {
			return f, err
		}
	}

	return f, nil
}

func parseDate(date string) (time.Time, error) {
	formats := []string{
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
		http.TimeFormat,
	}

	date = strings.TrimSpace(date)
	var err error
	var t time.Time
	for _, f := range formats {
		t, err = time.Parse(f, date)
		if err == nil {
			return t, nil
		}
	}

	return t, err
}
