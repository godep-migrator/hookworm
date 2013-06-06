package hookworm

import (
	"regexp"
	"strings"
)

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
