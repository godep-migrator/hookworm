package hookworm

import (
	"testing"
)

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
