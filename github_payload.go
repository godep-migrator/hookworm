package hookworm

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var pullRequestMessageRe = regexp.MustCompile("Merge pull request #[0-9]+ from.*")

type GithubPayload struct {
	Ref        *NullableString `json:"ref"`
	After      *NullableString `json:"after"`
	Before     *NullableString `json:"before"`
	Created    *NullableBool   `json:"created"`
	Deleted    *NullableBool   `json:"deleted"`
	Forced     *NullableBool   `json:"forced"`
	Compare    *NullableString `json:"compare"`
	Commits    []*Commit       `json:"commits"`
	HeadCommit *Commit         `json:"head_commit"`
	Repository *Repository     `json:"repository"`
	Pusher     *Pusher         `json:"pusher"`
}

/*
	TODO assign this at `GithubPayload` parse time
	so that we can serialize the payload as JSON for each ShellHandler instance
*/
func (me *GithubPayload) IsPullRequestMerge() bool {
	return len(me.Commits) > 1 &&
		pullRequestMessageRe.Match([]byte(me.HeadCommit.Message.String()))
}

/*
	TODO assign this at `GithubPayload` parse time
*/
func (me *GithubPayload) Paths() []string {
	var (
		paths   []string
		commits []*Commit
	)

	for _, commit := range me.Commits {
		commits = append(commits, commit)
	}

	commits = append(commits, me.HeadCommit)

	for i, commit := range commits {
		if me.IsPullRequestMerge() && i == 0 {
			continue
		}
		for _, path := range commit.Paths() {
			paths = append(paths, path)
		}
	}
	return paths
}

/*
	TODO assign this at `GithubPayload` parse time
*/
func (me *GithubPayload) IsValid() bool {
	return me.Ref != nil && me.After != nil &&
		me.Before != nil && me.Created != nil &&
		me.Deleted != nil && me.Forced != nil &&
		me.Compare != nil && me.Commits != nil &&
		me.Repository != nil && me.Pusher != nil
}

type Commit struct {
	Id        *NullableString   `json:"id"`
	Distinct  *NullableBool     `json:"distinct"`
	Message   *NullableString   `json:"message"`
	Timestamp *NullableString   `json:"timestamp"`
	Url       *NullableString   `json:"url"`
	Author    *Author           `json:"author"`
	Committer *Author           `json:"committer"`
	Added     []*NullableString `json:"added"`
	Removed   []*NullableString `json:"removed"`
	Modified  []*NullableString `json:"modified"`
}

func (me *Commit) Paths() []string {
	var paths []string
	pathSet := make(map[string]bool)

	for _, pathList := range [][]*NullableString{me.Added, me.Removed, me.Modified} {
		for _, path := range pathList {
			pathSet[path.String()] = true
		}
	}

	for path, _ := range pathSet {
		paths = append(paths, path)
	}

	sort.Strings(paths)
	return paths
}

type Repository struct {
	Id           *NullableInt64  `json:"id"`
	Name         *NullableString `json:"name"`
	Url          *NullableString `json:"url"`
	Description  *NullableString `json:"description"`
	Watchers     *NullableInt64  `json:"watchers"`
	Stargazers   *NullableInt64  `json:"stargazers"`
	Forks        *NullableInt64  `json:"forks"`
	Fork         *NullableBool   `json:"fork"`
	Size         *NullableInt64  `json:"size"`
	Owner        *Owner          `json:"owner"`
	Private      *NullableBool   `json:"private"`
	OpenIssues   *NullableInt64  `json:"open_issues"`
	HasIssues    *NullableBool   `json:"has_issues"`
	HasDownloads *NullableBool   `json:"has_downloads"`
	HasWiki      *NullableBool   `json:"has_wiki"`
	Language     *NullableString `json:"language"`
	CreatedAt    *NullableInt64  `json:"created_at"`
	PushedAt     *NullableInt64  `json:"pushed_at"`
	MasterBranch *NullableString `json:"master_branch"`
	Organization *NullableString `json:"organization"`
}

type Pusher struct {
	Name *NullableString `json:"name"`
}

type Author struct {
	Name     *NullableString `json:"name"`
	Email    *NullableString `json:"email"`
	Username *NullableString `json:"username"`
}

type Owner struct {
	Name  *NullableString `json:"name"`
	Email *NullableString `json:"email"`
}

type NullableString struct {
	Value  string
	isNull bool
}

func (me *NullableString) UnmarshalJSON(raw []byte) error {
	if bytes.Equal(raw, []byte("null")) {
		me.isNull = true
		return nil
	}
	me.Value = strings.TrimRight(strings.TrimLeft(string(raw), "\""), "\"")
	return nil
}

func (me *NullableString) MarshalJSON() ([]byte, error) {
	if me.isNull {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("%q", me.Value)), nil
}

func (me *NullableString) String() string {
	str := string(me.Value)
	str = strings.Replace(str, "\\n", "\n", -1)
	str = strings.Replace(str, "\\t", "\t", -1)
	return str
}

func (me *NullableString) Html() string {
	str := string(me.Value)
	str = strings.Replace(str, "\\n", "<br/>", -1)
	str = strings.Replace(str, "\\t", "    ", -1)
	return str
}

type NullableInt64 struct {
	Value  int64
	isNull bool
}

func (me *NullableInt64) UnmarshalJSON(raw []byte) error {
	if bytes.Equal(raw, []byte("null")) {
		me.isNull = true
		return nil
	}
	value, err := strconv.ParseInt(string(raw), 10, 64)
	if err != nil {
		me.isNull = true
		return err
	}
	me.Value = value
	return nil
}

func (me *NullableInt64) MarshalJSON() ([]byte, error) {
	if me.isNull {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatInt(me.Value, 10)), nil
}

func (me *NullableInt64) String() string {
	return strconv.FormatInt(me.Value, 10)
}

type NullableBool struct {
	Value  bool
	isNull bool
}

func (me *NullableBool) UnmarshalJSON(raw []byte) error {
	if bytes.Equal(raw, []byte("null")) {
		me.isNull = true
		return nil
	}
	value, err := strconv.ParseBool(string(raw))
	if err != nil {
		me.isNull = true
		return err
	}
	me.Value = value
	return nil
}

func (me *NullableBool) MarshalJSON() ([]byte, error) {
	if me.isNull {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatBool(me.Value)), nil
}

func (me *NullableBool) String() string {
	return strconv.FormatBool(me.Value)
}
