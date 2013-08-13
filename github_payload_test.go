package hookworm_test

import (
	"testing"

	. "hookworm"
)

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
