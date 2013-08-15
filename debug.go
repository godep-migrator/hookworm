// +build debug full

package hookworm

import (
	"expvar"
	// expose this when building with debug tag
	_ "net/http/pprof"
)

func init() {
	expvar.Publish("version", expvar.Func(expvarVersion))
	expvar.Publish("revision", expvar.Func(expvarRevision))
	expvar.Publish("build_tags", expvar.Func(expvarBuildTags))
}

func expvarVersion() interface{} {
	return VersionString
}

func expvarRevision() interface{} {
	return RevisionString
}

func expvarBuildTags() interface{} {
	return BuildTags
}
