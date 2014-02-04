package hookworm

import (
	"testing"
)

func TestExitNoopError(t *testing.T) {
	s := (&exitNoop{}).Error()
	if s != "exit noop 78" {
		t.Fail()
	}
}
