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
func TestWormFlagMapSetIgnoresEmptyishValues(t *testing.T) {
	wfm := newWormFlagMap()

	wfm.Set("")
	if wfm.String() != "" {
		t.Fail()
	}

	wfm.Set("      ")
	if wfm.String() != "" {
		t.Fail()
	}

	wfm.Set("\t\t\t\n\n\n\t\n   ")
	if wfm.String() != "" {
		t.Fail()
	}
}

func TestWormFlagMapMarshalJSON(t *testing.T) {
	wfm := newWormFlagMap()
	wfm.Set("fizz=buzz")

	json, err := wfm.MarshalJSON()
	if err != nil {
		t.Error(err)
	}

	if string(json) != `{"fizz":"buzz"}` {
		t.Fail()
	}
}

func TestWormFlagMapUnmarshalJSON(t *testing.T) {
	wfm := newWormFlagMap()
	err := wfm.UnmarshalJSON([]byte(`{"ham":"bone"}`))
	if err != nil {
		t.Error(err)
	}

	val := wfm.Get("ham").(string)
	if val != "bone" {
		t.Fail()
	}
}

func TestWormFlagMapGet(t *testing.T) {
	wfm := newWormFlagMap()
	wfm.Set("wat=herp")
	if wfm.Get("wat").(string) != "herp" {
		t.Fail()
	}
	if wfm.Get("nope").(string) != "" {
		t.Fail()
	}
}
