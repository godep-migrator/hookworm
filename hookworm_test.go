package hookworm_test

import (
	"testing"

	"github.com/modcloth-labs/hookworm"
)

func TestNullableStringTextHasEscapedTabs(t *testing.T) {
	s := &hookworm.NullableString{Value: "foo\\tbar"}
	if s.String() != "foo\tbar" {
		t.Fail()
	}
}

func TestNullableStringHtmlHasEscapedTabs(t *testing.T) {
	s := &hookworm.NullableString{Value: "foo\\tbar"}
	if s.Html() != "foo    bar" {
		t.Fail()
	}
}
