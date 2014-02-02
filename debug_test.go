package hookworm

import (
	"testing"
)

func TestExpvarVersion(t *testing.T) {
	v := expvarVersion()
	if v == nil {
		t.Fail()
	}
	switch v.(type) {
	case string:
		return
	default:
		t.Fail()
	}
}

func TestExpvarRevision(t *testing.T) {
	v := expvarRevision()
	if v == nil {
		t.Fail()
	}
	switch v.(type) {
	case string:
		return
	default:
		t.Fail()
	}
}

func TestExpvarBuildTags(t *testing.T) {
	v := expvarBuildTags()
	if v == nil {
		t.Fail()
	}
	switch v.(type) {
	case string:
		return
	default:
		t.Fail()
	}
}
