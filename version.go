package hookworm

import (
	"fmt"
	"os"
	"path"
)

var (
	// VersionString is the git description set via ldflags
	VersionString string
	// RevisionString is the git revision set via ldflags
	RevisionString string
	// BuildTags are the tags used at build time set via ldflags
	BuildTags string
	progName  string
)

func init() {
	progName = path.Base(os.Args[0])
	if RevisionString == "" {
		RevisionString = "<unknown>"
	}
	if VersionString == "" {
		VersionString = "<unknown>"
	}
}

func printVersion() {
	fmt.Println(progVersion())
}

func printRevision() {
	fmt.Println(RevisionString)
}

func progVersion() string {
	return fmt.Sprintf("%s %s", progName, VersionString)
}

func printVersionRevTags() {
	fmt.Printf("%s\nrev: %s\ntags: %s\n", progVersion(), RevisionString, BuildTags)
}
