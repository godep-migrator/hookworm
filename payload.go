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

type Payload struct {
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

func (me *Payload) IsPullRequestMerge() bool {
	return len(me.Commits) > 1 &&
		pullRequestMessageRe.Match([]byte(me.HeadCommit.Message.String()))
}

func (me *Payload) Paths() []string {
	var (
		paths   []string
		commits []*Commit
	)
	commits = append(commits, me.Commits...)
	commits = append(commits, me.HeadCommit)

	for i, commit := range me.Commits {
		if me.IsPullRequestMerge() && i == 0 {
			continue
		}
		paths = append(paths, commit.Paths()...)
	}
	return paths
}

func (me *Payload) IsValid() bool {
	return me.Ref != nil && me.After != nil &&
		me.Before != nil && me.Created != nil &&
		me.Deleted != nil && me.Forced != nil &&
		me.Compare != nil && me.Commits != nil &&
		me.HeadCommit != nil && me.Repository != nil &&
		me.Pusher != nil
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
	var pathKeys []string
	paths := make(map[string]bool)

	for _, pathList := range [][]*NullableString{me.Added, me.Removed, me.Modified} {
		for _, path := range pathList {
			paths[path.String()] = true
		}
	}

	for path, _ := range paths {
		pathKeys = append(pathKeys, path)
	}

	sort.Strings(pathKeys)
	return pathKeys
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
	value  string
	isNull bool
}

func (me *NullableString) UnmarshalJSON(raw []byte) error {
	if bytes.Equal(raw, []byte("null")) {
		me.isNull = true
		return nil
	}
	me.value = strings.TrimRight(strings.TrimLeft(string(raw), "\""), "\"")
	return nil
}

func (me *NullableString) MarshalJSON() ([]byte, error) {
	if me.isNull {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("%q", me.value)), nil
}

func (me *NullableString) String() string {
	return strings.Replace(string(me.value), "\\n", "\n", -1)
}

func (me *NullableString) Html() string {
	return strings.Replace(string(me.value), "\\n", "<br/>", -1)
}

type NullableInt64 struct {
	value  int64
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
	me.value = value
	return nil
}

func (me *NullableInt64) MarshalJSON() ([]byte, error) {
	if me.isNull {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatInt(me.value, 10)), nil
}

func (me *NullableInt64) String() string {
	return strconv.FormatInt(me.value, 10)
}

type NullableBool struct {
	value  bool
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
	me.value = value
	return nil
}

func (me *NullableBool) MarshalJSON() ([]byte, error) {
	if me.isNull {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatBool(me.value)), nil
}

func (me *NullableBool) String() string {
	return strconv.FormatBool(me.value)
}
