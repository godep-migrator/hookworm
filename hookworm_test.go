package hookworm_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"

	. "hookworm"
)

var (
	here = ""
)

func init() {
	_, filename, _, _ := runtime.Caller(1)
	here = path.Dir(filename)
}

func getRawPayloadJSON(name, subdir string) []byte {
	filename := path.Join(here, "sampledata", subdir,
		fmt.Sprintf("%s.json", name))
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	jsonBytes, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	return jsonBytes
}

func getGithubPayload(name string) *GithubPayload {
	payload := &GithubPayload{}

	jsonBytes := getRawPayloadJSON(name, "github-payloads")
	err := json.Unmarshal(jsonBytes, payload)
	if err != nil {
		panic(err)
	}

	return payload
}

func getTravisPayload(name string) *TravisPayload {
	payload := &TravisPayload{}

	jsonBytes := getRawPayloadJSON(name, "travis-payloads")
	err := json.Unmarshal(jsonBytes, payload)
	if err != nil {
		panic(err)
	}

	return payload
}
