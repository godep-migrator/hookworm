package hookworm

import (
	"testing"
)

var (
	ctypesAbbrs = map[string]string{
		"application/x-www-form-urlencoded; charset=utf-8": "application/x-www-form-urlencoded",
		"text/plain; charset=utf-8":                        "text/plain",
		"text/javascript     ; charset=utf-8":              "text/javascript",
		"Application/JSON":                                 "application/json",
	}
)

func TestAbbreviatedContentType(t *testing.T) {
	for before, after := range ctypesAbbrs {
		if abbrCtype(before) != after {
			t.Fail()
		}
	}
}
