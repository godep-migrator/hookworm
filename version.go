package hookworm

import (
	"fmt"
	"os"
	"path"
)

var (
	VersionString  string
	RevisionString string
	progName       string
)

func init() {
	progName = path.Base(os.Args[0])
}

func printVersion() {
	fmt.Println(progVersion())
}

func printRevision() {
	if RevisionString == "" {
		RevisionString = "<unknown>"
	}
	fmt.Println(RevisionString)
}

func progVersion() string {
	if VersionString == "" {
		VersionString = "<unknown>"
	}

	return fmt.Sprintf("%s %s", progName, VersionString)
}
