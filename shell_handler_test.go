package hookworm

import (
	"os"
	"path"
	"strings"
	"testing"
)

const noopHandlerBody = `#!/usr/bin/env python
import sys

sys.stdout.write(sys.stdin.read())
sys.exit(0)
`

var (
	noopHandlerPath = ""

	shellHandlerConfig = &HandlerConfig{
		WormTimeout: 5,
		Debug:       true,
	}
)

func init() {
	f, err := os.Create(path.Join(os.TempDir(), "noop-handler.py"))
	if err != nil {
		panic(err)
	}
	noopHandlerPath = f.Name()
	defer f.Close()
	f.WriteString(noopHandlerBody)
	f.Chmod(0755)
}

func setupShellHandler(t *testing.T) *shellHandler {
	sh, err := newShellHandler(noopHandlerPath, shellHandlerConfig)
	if err != nil {
		t.Error(err)
	}
	return sh
}

func assertNoopWorks(out string, err error, t *testing.T) {
	if err != nil {
		t.Errorf("noop shell handler error: %v", err)
	}

	if strings.TrimSpace(out) != `{}` {
		t.Fail()
	}
}

func TestShellHandlerHandleGithubPayload(t *testing.T) {
	sh := setupShellHandler(t)
	out, err := sh.HandleGithubPayload(`{}`)
	assertNoopWorks(out, err, t)
}

func TestShellHandlerHandleTravisPayload(t *testing.T) {
	sh := setupShellHandler(t)
	out, err := sh.HandleTravisPayload(`{}`)
	assertNoopWorks(out, err, t)
}
