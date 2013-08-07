package hookworm_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"testing"

	. "hookworm"
)

var (
	here = ""
)

func init() {
	_, filename, _, _ := runtime.Caller(1)
	here = path.Dir(filename)
}

func getGithubPayload(name string) *GithubPayload {
	payload := &GithubPayload{}

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

func TestNullableStringTextHasEscapedNewlines(t *testing.T) {
	s := &NullableString{Value: "foo\\nbar"}
	if s.String() != "foo\nbar" {
		t.Fail()
	}
}

func TestNullableStringHtmlHasEscapedNewlines(t *testing.T) {
	s := &NullableString{Value: "foo\\nbar"}
	if s.Html() != "foo<br/>bar" {
		t.Fail()
	}
}

func TestGithubPayloadDetectsPullRequests(t *testing.T) {
	payload := getGithubPayload("pull_request")
	if !payload.IsPullRequestMerge() {
		t.Fail()
	}
}

func TestGithubPayloadDetectsNonPullRequest(t *testing.T) {
	payload := getGithubPayload("valid")
	if payload.IsPullRequestMerge() {
		t.Fail()
	}
}
