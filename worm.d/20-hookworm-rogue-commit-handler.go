package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/smtp"
	"net/url"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var (
	rogueCommitEmailTmpl = template.Must(template.New("email").Parse(`From: {{.From}}
To: {{.Recipients}}
Subject: [hookworm] Rogue commit by {{.HeadCommitAuthor}} to {{.Repo}} {{.Ref}} ({{.HeadCommitID}})
Date: {{.Date}}
Message-ID: <{{.MessageID}}@{{.Hostname}}>
List-ID: {{.Repo}} <hookworm.github.com>
Content-Type: multipart/alternative;
  boundary="--==ZOMGBOUNDARAAAYYYYY";
  charset=UTF-8
Content-Transfer-Encoding: 7bit

----==ZOMGBOUNDARAAAYYYYY
Date: {{.Date}}
Mime-Version: 1.0
Content-Type: text/plain; charset=utf8
Content-Transfer-Encoding: 7bit

Rogue commit detected!

Repo      {{.Repo}}
Ref       {{.Ref}}
ID        {{.HeadCommitURL}}
Author    {{.HeadCommitAuthor}}
Committer {{.HeadCommitCommitter}}
Timestamp {{.HeadCommitTimestamp}}
Message   {{.HeadCommitMessageText}}

-- 
This email was sent by hookworm:  https://github.com/modcloth-labs/hookworm

A rogue commit is a commit made directly to a branch that is being watched such
that only pull requests should be merged into them.

The configured watched branches are:

{{range .WatchedBranches}} - {{.}}
{{end}}

The configured watched paths are:

{{range .WatchedPaths}} - {{.}}
{{end}}


If you believe this rogue commit email is an error, you should hunt down the
party responsible for the hookworm instance registered as a WebHook URL in this
repo's service hook settings ({{.RepoURL}}/settings/hooks).

Pretty please submit issues specific to hookworm functionality on github:
https://github.com/modcloth-labs/hookworm/issues/new

----==ZOMGBOUNDARAAAYYYYY
Date: {{.Date}}
Mime-Version: 1.0
Content-Type: text/html; charset=utf8
Content-Transfer-Encoding: 7bit

<div>
  <h1><a href="{{.HeadCommitURL}}">Rogue commit detected!</a></h1>

  <table>
    <thead><th></th><th></th></thead>
    <tbody>
      <tr>
        <td style="text-align:right;vertical-align:top;">
          <strong>Repo</strong>:
        </td>
        <td>{{.Repo}}</td>
      </tr>
      <tr>
        <td style="text-align:right;vertical-align:top;">
          <strong>Ref</strong>:
        </td>
        <td>{{.Ref}}</td>
      </tr>
      <tr>
        <td style="text-align:right;vertical-align:top;">
          <strong>ID</strong>:
        </td>
        <td><a href="{{.HeadCommitURL}}">{{.HeadCommitID}}</a></td>
      </tr>
      <tr>
        <td style="text-align:right;vertical-align:top;">
          <strong>Author</strong>:
        </td>
        <td>{{.HeadCommitAuthor}}</td>
      </tr>
      <tr>
        <td style="text-align:right;vertical-align:top;">
          <strong>Committer</strong>:
        </td>
        <td>{{.HeadCommitCommitter}}</td>
      </tr>
      <tr>
        <td style="text-align:right;vertical-align:top;">
          <strong>Timestamp</strong>:
        </td>
        <td>{{.HeadCommitTimestamp}}</td>
      </tr>
      <tr>
        <td style="text-align:right;vertical-align:top;">
          <strong>Message</strong>:
        </td>
        <td>{{.HeadCommitMessageHTML}}</td>
      </tr>
    </tbody>
  </table>
</div>

<hr/>
<div style="font-size:.8em">
  <p>
    This email was sent by
    <a href="https://github.com/modcloth-labs/hookworm">hookworm</a>.
  </p>

  <p>
    A rogue commit is a commit made directly to a branch that is being watched such
    that only pull requests should be merged into them.
  </p>

  <p>
    The configured watched branches are:
  </p>

  <ul>
    {{range .WatchedBranches}}<li><strong>{{.}}</strong></li>
    {{end}}
  </ul>

  <p>
    The configured watched paths are:
  </p>

  <ul>
    {{range .WatchedPaths}}<li><strong>{{.}}</strong></li>
    {{end}}
  </ul>

  <p>
    If you believe this rogue commit email is an error, you should hunt down the
    party responsible for the hookworm instance registered as a WebHook URL in
    this repo's <a href="{{.RepoURL}}/settings/hooks">service hook settings</a>.
  </p>

  <p>
    Pretty please submit issues specific to hookworm functionality
    <a href="https://github.com/modcloth-labs/hookworm/issues/new">on github</a>
  </p>

</div>

----==ZOMGBOUNDARAAAYYYYY--
`))

	hostname string

	rfc2822DateFmt = "Mon, 02 Jan 2006 15:04:05 -0700"
)

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		hostname = "somewhere.local"
	}
}

type travisPayload struct {
	ID            int               `json:"id"`
	Repository    *travisRepository `json:"repository"`
	Number        string            `json:"number"`
	Config        interface{}       `json:"config"`
	Status        int               `json:"status"`
	Result        int               `json:"result"`
	StatusMessage string            `json:"status_message"`
	ResultMessage string            `json:"result_message"`
	StartedAt     *time.Time        `json:"started_at"`
	FinishedAt    *time.Time        `json:"finished_at"`
	Duration      int               `json:"duration"`
}

type travisRepository struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	OwnerName string `json:"owner_name"`
	URL       string `json:"url"`
}

type githubPayload struct {
	Ref        *nullableString `json:"ref"`
	After      *nullableString `json:"after"`
	Before     *nullableString `json:"before"`
	Created    *nullableBool   `json:"created"`
	Deleted    *nullableBool   `json:"deleted"`
	Forced     *nullableBool   `json:"forced"`
	Compare    *nullableString `json:"compare"`
	Commits    []*commit       `json:"commits"`
	HeadCommit *commit         `json:"head_commit"`
	Repository *repository     `json:"repository"`
	Pusher     *pusher         `json:"pusher"`
	IsPrMerge  *nullableBool   `json:"is_pr_merge"`
}

func (ghp *githubPayload) Paths() []string {
	var (
		paths   []string
		commits []*commit
	)

	for _, commit := range ghp.Commits {
		commits = append(commits, commit)
	}

	commits = append(commits, ghp.HeadCommit)

	for i, commit := range commits {
		if ghp.IsPrMerge.Value && i == 0 {
			continue
		}
		for _, path := range commit.Paths() {
			paths = append(paths, path)
		}
	}
	return paths
}

func (ghp *githubPayload) IsValid() bool {
	return ghp.Ref != nil && ghp.After != nil &&
		ghp.Before != nil && ghp.Created != nil &&
		ghp.Deleted != nil && ghp.Forced != nil &&
		ghp.Compare != nil && ghp.Commits != nil &&
		ghp.Repository != nil && ghp.Pusher != nil
}

type commit struct {
	ID        *nullableString   `json:"id"`
	Distinct  *nullableBool     `json:"distinct"`
	Message   *nullableString   `json:"message"`
	Timestamp *nullableString   `json:"timestamp"`
	URL       *nullableString   `json:"url"`
	Author    *author           `json:"author"`
	Committer *author           `json:"committer"`
	Added     []*nullableString `json:"added"`
	Removed   []*nullableString `json:"removed"`
	Modified  []*nullableString `json:"modified"`
}

func (c *commit) Paths() []string {
	var paths []string
	pathSet := make(map[string]bool)

	for _, pathList := range [][]*nullableString{c.Added, c.Removed, c.Modified} {
		for _, path := range pathList {
			pathSet[path.String()] = true
		}
	}

	for path := range pathSet {
		paths = append(paths, path)
	}

	sort.Strings(paths)
	return paths
}

type repository struct {
	ID           *nullableInt64  `json:"id"`
	Name         *nullableString `json:"name"`
	URL          *nullableString `json:"url"`
	Description  *nullableString `json:"description"`
	Watchers     *nullableInt64  `json:"watchers"`
	Stargazers   *nullableInt64  `json:"stargazers"`
	Forks        *nullableInt64  `json:"forks"`
	Fork         *nullableBool   `json:"fork"`
	Size         *nullableInt64  `json:"size"`
	Owner        *owner          `json:"owner"`
	Private      *nullableBool   `json:"private"`
	OpenIssues   *nullableInt64  `json:"open_issues"`
	HasIssues    *nullableBool   `json:"has_issues"`
	HasDownloads *nullableBool   `json:"has_downloads"`
	HasWiki      *nullableBool   `json:"has_wiki"`
	Language     *nullableString `json:"language"`
	CreatedAt    *nullableInt64  `json:"created_at"`
	PushedAt     *nullableInt64  `json:"pushed_at"`
	MasterBranch *nullableString `json:"master_branch"`
	Organization *nullableString `json:"organization"`
}

type pusher struct {
	Name *nullableString `json:"name"`
}

type author struct {
	Name     *nullableString `json:"name"`
	Email    *nullableString `json:"email"`
	Username *nullableString `json:"username"`
}

type owner struct {
	Name  *nullableString `json:"name"`
	Email *nullableString `json:"email"`
}

type nullableString struct {
	Value  string
	isNull bool
}

func (s *nullableString) UnmarshalJSON(raw []byte) error {
	if bytes.Equal(raw, []byte("null")) {
		s.isNull = true
		return nil
	}
	s.Value = strings.TrimRight(strings.TrimLeft(string(raw), "\""), "\"")
	return nil
}

func (s *nullableString) MarshalJSON() ([]byte, error) {
	if s.isNull {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("%q", s.Value)), nil
}

func (s *nullableString) String() string {
	str := string(s.Value)
	str = strings.Replace(str, "\\n", "\n", -1)
	str = strings.Replace(str, "\\t", "\t", -1)
	return str
}

func (s *nullableString) HTML() string {
	str := string(s.Value)
	str = strings.Replace(str, "\\n", "<br/>", -1)
	str = strings.Replace(str, "\\t", "    ", -1)
	return str
}

type nullableInt64 struct {
	Value  int64
	isNull bool
}

func (i *nullableInt64) UnmarshalJSON(raw []byte) error {
	if bytes.Equal(raw, []byte("null")) {
		i.isNull = true
		return nil
	}
	value, err := strconv.ParseInt(string(raw), 10, 64)
	if err != nil {
		i.isNull = true
		return err
	}
	i.Value = value
	return nil
}

func (i *nullableInt64) MarshalJSON() ([]byte, error) {
	if i.isNull {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatInt(i.Value, 10)), nil
}

func (i *nullableInt64) String() string {
	return strconv.FormatInt(i.Value, 10)
}

type nullableBool struct {
	Value  bool
	isNull bool
}

func (b *nullableBool) UnmarshalJSON(raw []byte) error {
	if bytes.Equal(raw, []byte("null")) {
		b.isNull = true
		return nil
	}
	value, err := strconv.ParseBool(string(raw))
	if err != nil {
		b.isNull = true
		return err
	}
	b.Value = value
	return nil
}

func (b *nullableBool) MarshalJSON() ([]byte, error) {
	if b.isNull {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatBool(b.Value)), nil
}

func (b *nullableBool) String() string {
	return strconv.FormatBool(b.Value)
}

type wormFlagMap struct {
	EmailFromAddr   string `json:"email_from_addr"`
	EmailRcpts      string `json:"email_recipients"`
	EmailUri        string `json:"email_uri"`
	WatchedBranches string `json:"watched_branches"`
	WatchedPaths    string `json:"watched_paths"`
}

type handlerConfig struct {
	Debug      bool         `json:"debug"`
	WorkingDir string       `json:"working_dir"`
	WormDir    string       `json:"worm_dir"`
	WormFlags  *wormFlagMap `json:"worm_flags"`
}

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

type rogueCommitHandler struct {
	debug                  bool
	emailer                *emailer
	fromAddr               string
	recipients             []string
	watchedBranches        []*regexp.Regexp
	watchedPaths           []*regexp.Regexp
	watchedBranchesStrings []string
	watchedPathsStrings    []string
}

type rogueCommitEmailContext struct {
	From                  string
	Recipients            string
	Date                  string
	MessageID             string
	Hostname              string
	Repo                  string
	Ref                   string
	WatchedBranches       []string
	WatchedPaths          []string
	RepoURL               string
	HeadCommitID          string
	HeadCommitURL         string
	HeadCommitAuthor      string
	HeadCommitCommitter   string
	HeadCommitMessageText string
	HeadCommitMessageHTML string
	HeadCommitTimestamp   string
}

func newRogueCommitHandler(cfg *handlerConfig) *rogueCommitHandler {
	handler := &rogueCommitHandler{
		debug:           cfg.Debug,
		emailer:         newEmailer(cfg.WormFlags.EmailUri),
		fromAddr:        cfg.WormFlags.EmailFromAddr,
		recipients:      commaSplit(cfg.WormFlags.EmailRcpts),
		watchedBranches: strsToRegexes(commaSplit(cfg.WormFlags.WatchedBranches)),
		watchedPaths:    strsToRegexes(commaSplit(cfg.WormFlags.WatchedPaths)),
	}

	for _, re := range handler.watchedBranches {
		handler.watchedBranchesStrings = append(handler.watchedBranchesStrings, re.String())
	}

	for _, re := range handler.watchedPaths {
		handler.watchedPathsStrings = append(handler.watchedPathsStrings, re.String())
	}

	return handler
}

func (rch *rogueCommitHandler) HandleGithubPayload(payload *githubPayload) error {
	if !rch.isWatchedBranch(payload.Ref.String()) {
		if rch.debug {
			log.Printf("%v is not a watched branch, yay!\n", payload.Ref.String())
		}
		return nil
	}

	if rch.debug {
		log.Printf("%v is a watched branch!\n", payload.Ref.String())
	}

	hcID := payload.HeadCommit.ID.String()

	if payload.IsPrMerge.Value {
		if rch.debug {
			log.Printf("%v is a pull request merge, yay!\n", hcID)
		}
		return nil
	}

	if rch.debug {
		log.Printf("%v is not a pull request merge!\n", hcID)
	}

	if !rch.hasWatchedPath(payload.Paths()) {
		if rch.debug {
			log.Printf("%v does not contain watched paths, yay!\n", hcID)
		}
		return nil
	}

	if rch.debug {
		log.Printf("%v contains watched paths!\n", hcID)
	}

	if err := rch.alert(payload); err != nil {
		if rch.debug {
			log.Printf("ERROR sending alert: %+v\n", err)
		}
		return err
	}

	if rch.debug {
		log.Printf("Sent alert to %+v\n", rch.recipients)
	}
	return nil
}

func (rch *rogueCommitHandler) HandleTravisPayload(*travisPayload) error {
	return nil
}

func (rch *rogueCommitHandler) isWatchedBranch(ref string) bool {
	sansRefsHeads := strings.Replace(ref, "refs/heads/", "", 1)
	for _, branchRe := range rch.watchedBranches {
		if branchRe.MatchString(sansRefsHeads) {
			return true
		}
	}
	return false
}

func (rch *rogueCommitHandler) hasWatchedPath(paths []string) bool {
	for _, pathRe := range rch.watchedPaths {
		for _, path := range paths {
			if pathRe.MatchString(path) {
				return true
			}
		}
	}
	return false
}

func (rch *rogueCommitHandler) alert(payload *githubPayload) error {
	log.Printf("WARNING rogue commit! %+v, head commit: %+v\n",
		payload, payload.HeadCommit)
	if len(rch.recipients) == 0 {
		log.Println("No email recipients specified, so no emailing!")
		return nil
	}

	hc := payload.HeadCommit
	ctx := &rogueCommitEmailContext{
		From:       rch.fromAddr,
		Recipients: strings.Join(rch.recipients, ", "),
		Date:       time.Now().UTC().Format(rfc2822DateFmt),
		MessageID:  fmt.Sprintf("%v", time.Now().UTC().UnixNano()),
		Hostname:   hostname,
		Repo: fmt.Sprintf("%s/%s", payload.Repository.Owner.Name.String(),
			payload.Repository.Name.String()),
		Ref:                   payload.Ref.String(),
		WatchedBranches:       rch.watchedBranchesStrings,
		WatchedPaths:          rch.watchedPathsStrings,
		RepoURL:               payload.Repository.URL.String(),
		HeadCommitID:          hc.ID.String(),
		HeadCommitURL:         hc.URL.String(),
		HeadCommitAuthor:      hc.Author.Name.String(),
		HeadCommitCommitter:   hc.Committer.Name.String(),
		HeadCommitMessageText: hc.Message.String(),
		HeadCommitMessageHTML: hc.Message.HTML(),
		HeadCommitTimestamp:   hc.Timestamp.String(),
	}
	var emailBuf bytes.Buffer

	err := rogueCommitEmailTmpl.Execute(&emailBuf, ctx)
	if err != nil {
		return err
	}

	if rch.debug {
		log.Printf("Email message:\n%v\n", string(emailBuf.Bytes()))
	}
	return rch.emailer.Send(rch.fromAddr, rch.recipients, emailBuf.Bytes())
}

type emailer struct {
	serverUri *url.URL
	auth      smtp.Auth
}

func newEmailer(serverUri string) *emailer {
	parsedUri, err := url.Parse(serverUri)
	if err != nil {
		log.Panicln("Failed to parse smtp url:", err)
	}

	var auth smtp.Auth

	if parsedUri.User != nil {
		username := parsedUri.User.Username()
		password, _ := parsedUri.User.Password()
		auth = smtp.PlainAuth("",
			username, password, parsedUri.Host)
	}

	return &emailer{serverUri: parsedUri, auth: auth}
}

func (em *emailer) Send(from string, to []string, msg []byte) error {
	return insecureSendMail(em.serverUri.Host, em.auth, from, to, msg)
}

func insecureSendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	config := &tls.Config{}
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	if ok, _ := c.Extension("STARTTLS"); ok {
		host, _, _ := net.SplitHostPort(addr)
		config.InsecureSkipVerify = true
		config.ServerName = host
		if err = c.StartTLS(config); err != nil {
			return err
		}
	}
	if a != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(a); err != nil {
				return err
			}
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

func configPath() (string, error) {
	workingDir := os.Getenv("HOOKWORM_WORKING_DIR")
	if len(workingDir) < 1 {
		return "", fmt.Errorf("missing HOOKWORM_WORKING_DIR")
	}

	return path.Join(workingDir, "20-hookworm-rogue-commit-handler.go.cfg.json"), nil
}

func config() (*handlerConfig, error) {
	stashedPath, err := configPath()
	if err != nil {
		return nil, err
	}

	cfg := &handlerConfig{}

	f, err := os.Open(stashedPath)
	if err != nil {
		return nil, err
	}

	rawCfg, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(rawCfg, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func configure() error {
	log.Printf("%s: configure!", path.Base(os.Args[0]))
	cfg := &handlerConfig{}

	rawConfig, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	err = json.Unmarshal(rawConfig, cfg)
	if err != nil {
		return err
	}

	stashedPath, err := configPath()
	if err != nil {
		return err
	}

	f, err := os.Create(stashedPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%s", string(rawConfig))
	if err != nil {
		return err
	}

	log.Printf("Wrote config to %v\n", stashedPath)

	return nil
}

func handleGithub() error {
	log.Printf("%s: handle github!", path.Base(os.Args[0]))

	payload := &githubPayload{}
	rawPayload, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(rawPayload, payload); err != nil {
		return err
	}

	if _, err := bytes.NewBuffer(rawPayload).WriteTo(os.Stdout); err != nil {
		return err
	}

	cfg, err := config()
	if err != nil {
		return err
	}

	handler := newRogueCommitHandler(cfg)
	return handler.HandleGithubPayload(payload)
}

func handleTravis() error {
	log.Printf("%s: handle travis!", path.Base(os.Args[0]))
	payload := &travisPayload{}
	rawPayload, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(rawPayload, payload); err != nil {
		return err
	}

	if _, err := bytes.NewBuffer(rawPayload).WriteTo(os.Stdout); err != nil {
		return err
	}

	cfg, err := config()
	if err != nil {
		return err
	}

	handler := newRogueCommitHandler(cfg)
	return handler.HandleTravisPayload(payload)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "ERROR: No command given")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "configure":
		if err := configure(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(86)
		}
		os.Exit(0)
	case "handle":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Handle what?")
			os.Exit(2)
		}
		switch os.Args[2] {
		case "github":
			if err := handleGithub(); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
				os.Exit(86)
			}
			os.Exit(0)
		case "travis":
			if err := handleTravis(); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
				os.Exit(86)
			}
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "ERROR: I don't know how to handle %q\n", os.Args[2])
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "ERROR: unknown command %q\n", os.Args[1])
		os.Exit(1)
	}
}
