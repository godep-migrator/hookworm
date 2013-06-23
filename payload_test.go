package hookworm_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"testing"

	. "github.com/modcloth-labs/hookworm"
)

var (
	here = ""
)

func init() {
	_, filename, _, _ := runtime.Caller(1)
	here = path.Dir(filename)
}

func getPayload(name string) *Payload {
	payload := &Payload{}

	filename := path.Join(here, "sampledata", "payloads",
		fmt.Sprintf("%s.json", name))
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	jsonBytes, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(jsonBytes, payload)
	if err != nil {
		panic(err)
	}

	return payload
}

func TestNullableStringTextHasEscapedTabs(t *testing.T) {
	s := &NullableString{Value: "foo\\tbar"}
	if s.String() != "foo\tbar" {
		t.Fail()
	}
}

func TestNullableStringHtmlHasEscapedTabs(t *testing.T) {
	s := &NullableString{Value: "foo\\tbar"}
	if s.Html() != "foo    bar" {
		t.Fail()
	}
}

func TestPayloadDetectsPullRequests(t *testing.T) {
	payload := getPayload("pull_request")
	if !payload.IsPullRequestMerge() {
		t.Fail()
	}
}

func TestPayloadDetectsNonPullRequest(t *testing.T) {
	payload := getPayload("valid")
	if payload.IsPullRequestMerge() {
		t.Fail()
	}
}
