package hookworm

import (
	"fmt"
	"io/ioutil"
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
	if len(workingDir) > 0 {
		fd, err := os.Create(filepath.Join(workingDir, ".write-test"))
		defer func() {
			if fd != nil {
				fd.Close()
			}
		}()

		if err != nil {
			return "", err
		}

		return workingDir, nil
	}

	return ioutil.TempDir("", fmt.Sprintf("hookworm-%d-", os.Getpid()))
}
