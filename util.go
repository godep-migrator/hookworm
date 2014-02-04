package hookworm

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	hostname string

	rfc2822DateFmt = "Mon, 02 Jan 2006 15:04:05 -0700"
)

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		hostname = "somewhere.local"
	}
}

func commaSplit(str string) []string {
	var ret []string

	for _, part := range strings.Split(str, ",") {
		part = strings.TrimSpace(part)
		if len(part) > 0 {
			ret = append(ret, part)
		}
	}

	return ret
}

func strsToRegexes(strs []string) []*regexp.Regexp {
	var regexps []*regexp.Regexp

	for _, str := range strs {
		regexps = append(regexps, regexp.MustCompile(str))
	}

	return regexps
}

func getWorkingDir(workingDir string) (string, error) {
	wd, err := getWriteableDir(workingDir, "")
	if err != nil {
		return "", err
	}

	if wd != "" {
		return wd, nil
	}

	tmpdir, err := ioutil.TempDir("", fmt.Sprintf("hookworm-%d-", os.Getpid()))
	if err != nil {
		return "", err
	}

	return tmpdir, nil
}

func getStaticDir(staticDir string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}

	return getWriteableDir(staticDir, filepath.Join(wd, "public"))
}

func getWriteableDir(dir, defaultDir string) (string, error) {
	if len(dir) > 0 {
		fd, err := os.Create(filepath.Join(dir, ".write-test"))
		defer func() {
			if fd != nil {
				fd.Close()
			}
		}()

		if err != nil {
			return "", err
		}

		return dir, nil
	}

	return defaultDir, nil
}

func extractPayload(l *hookwormLogger, r *http.Request) (string, error) {
	rawPayload := ""
	ctype := abbrCtype(r.Header.Get("Content-Type"))

	switch ctype {
	case "application/json", "text/javascript", "text/plain":
		rawPayloadBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return "", err
		}
		rawPayload = string(rawPayloadBytes)
	case "application/x-www-form-urlencoded":
		rawPayload = r.FormValue("payload")
	}

	if len(rawPayload) < 1 {
		l.Println("Empty payload!")
		return "", fmt.Errorf("empty payload")
	}

	l.Debugf("Raw payload: %+v\n", rawPayload)
	return rawPayload, nil
}

func abbrCtype(ctype string) string {
	s := strings.Split(ctype, ";")[0]
	return strings.ToLower(strings.TrimSpace(s))
}
