package hookworm

import (
	"regexp"
	"testing"
)

func TestNewWormFlagsMap(t *testing.T) {
	m := newWormFlagMap()
	if m == nil {
		t.Fail()
	}
	if m.values == nil {
		t.Fail()
	}
}

func TestWormFlagMapString(t *testing.T) {
	m := newWormFlagMap()
	m.Set("baz=bar; ham=true; qwwx=no; derp")
	s := m.String()
	if s == "" {
		t.Fail()
	}
	if ok, _ := regexp.MatchString("baz=bar;", s); !ok {
		t.Fail()
	}
	if ok, _ := regexp.MatchString("ham=true;", s); !ok {
		t.Fail()
	}
	if ok, _ := regexp.MatchString("derp=true;", s); !ok {
		t.Fail()
	}
}
