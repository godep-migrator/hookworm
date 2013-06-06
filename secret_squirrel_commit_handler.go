package hookworm

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"
)

var (
	rfc2822DateFmt        = "Mon, 02 Jan 2006 15:04:05 -0700"
	hostname              string
	secretCommitEmailTmpl = template.Must(template.New("email").Parse(`From: {{.From}}
To: {{.Recipients}}
Subject: [hookworm] Secret commit by {{.HeadCommitAuthor}} to {{.Repo}} {{.Ref}} ({{.HeadCommitId}})
Date: {{.Date}}
Message-ID: <{{.MessageId}}@{{.Hostname}}>
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

Secret commit detected!

Repo      {{.Repo}}
Ref       {{.Ref}}
Id        {{.HeadCommitUrl}}
Message   {{.HeadCommitMessage}}
Author    {{.HeadCommitAuthor}}
Committer {{.HeadCommitCommitter}}
Timestamp {{.HeadCommitTimestamp}}


----==ZOMGBOUNDARAAAYYYYY
Date: {{.Date}}
Mime-Version: 1.0
Content-Type: text/html; charset=utf8
Content-Transfer-Encoding: 7bit

<div>
  <h1><a href="{{.HeadCommitUrl}}">Secret commit detected!</a></h1>

  <table>
    <thead><th></th><th></th></thead>
    <tbody>
      <tr><td style="text-align:right;"><strong>Repo</strong>:</td><td>{{.Repo}}</td></tr>
      <tr><td style="text-align:right;"><strong>Ref</strong>:</td><td>{{.Ref}}</td></tr>
      <tr><td style="text-align:right;"><strong>Id</strong>:</td><td><a href="{{.HeadCommitUrl}}">{{.HeadCommitId}}</a></td></tr>
      <tr><td style="text-align:right;"><strong>Message</strong>:</td><td>{{.HeadCommitMessage}}</td></tr>
      <tr><td style="text-align:right;"><strong>Author</strong>:</td><td>{{.HeadCommitAuthor}}</td></tr>
      <tr><td style="text-align:right;"><strong>Committer</strong>:</td><td>{{.HeadCommitCommitter}}</td></tr>
      <tr><td style="text-align:right;"><strong>Timestamp</strong>:</td><td>{{.HeadCommitTimestamp}}</td></tr>
    </tbody>
  </table>
</div>

----==ZOMGBOUNDARAAAYYYYY--
`))
)

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		hostname = "somewhere.local"
	}
}

type SecretSquirrelCommitHandler struct {
	debug           bool
	emailer         *Emailer
	fromAddr        string
	recipients      []string
	policedBranches []*regexp.Regexp
	nextHandler     Handler
}

type secretCommitEmailContext struct {
	From                string
	Recipients          string
	Date                string
	MessageId           string
	Hostname            string
	Repo                string
	Ref                 string
	HeadCommitId        string
	HeadCommitUrl       string
	HeadCommitAuthor    string
	HeadCommitCommitter string
	HeadCommitMessage   string
	HeadCommitTimestamp string
}

func NewSecretSquirrelCommitHandler(cfg *HandlerConfig) *SecretSquirrelCommitHandler {
	handler := &SecretSquirrelCommitHandler{
		debug:           cfg.Debug,
		emailer:         NewEmailer(cfg.EmailUri),
		fromAddr:        cfg.EmailFromAddr,
		recipients:      cfg.EmailRcpts,
		policedBranches: strsToRegexes(cfg.PolicedBranches),
	}
	return handler
}

func (me *SecretSquirrelCommitHandler) HandlePayload(payload *Payload) error {
	if !me.isPolicedBranch(payload.Ref.String()) {
		if me.debug {
			log.Printf("%v is not a policed branch, yay!\n", payload.Ref.String())
		}
		return nil
	}

	if me.debug {
		log.Printf("%v is a policed branch!\n", payload.Ref.String())
	}

	if payload.IsPullRequestMerge() {
		if me.debug {
			log.Printf("%v is a pull request merge, yay!\n", payload.HeadCommit.Id.String())
		}
		return nil
	}

	if me.debug {
		log.Printf("%v is not a pull request merge!\n", payload.HeadCommit.Id.String())
	}

	if err := me.alert(payload); err != nil {
		if me.debug {
			log.Printf("ERROR sending alert: %+v\n", err)
		}
		return err
	}

	if me.debug {
		log.Printf("Sent alert to %+v\n", me.recipients)
	}
	return nil
}

func (me *SecretSquirrelCommitHandler) SetNextHandler(handler Handler) {
	me.nextHandler = handler
}

func (me *SecretSquirrelCommitHandler) NextHandler() Handler {
	return me.nextHandler
}

func (me *SecretSquirrelCommitHandler) isPolicedBranch(ref string) bool {
	sansRefsHeads := strings.Replace(ref, "refs/heads/", "", 1)
	for _, ref := range me.policedBranches {
		if ref.MatchString(sansRefsHeads) {
			return true
		}
	}
	return false
}

func (me *SecretSquirrelCommitHandler) alert(payload *Payload) error {
	log.Printf("WARNING secret squirrel commit! %+v, head commit: %+v\n",
		payload, payload.HeadCommit)
	if len(me.recipients) == 0 {
		log.Println("No email recipients specified, so no emailing!")
		return nil
	}

	hc := payload.HeadCommit
	ctx := &secretCommitEmailContext{
		From:       me.fromAddr,
		Recipients: strings.Join(me.recipients, ", "),
		Date:       time.Now().UTC().Format(rfc2822DateFmt),
		MessageId:  fmt.Sprintf("%v", time.Now().UTC().UnixNano()),
		Hostname:   hostname,
		Repo: fmt.Sprintf("%s/%s", payload.Repository.Owner.Name.String(),
			payload.Repository.Name.String()),
		Ref:                 payload.Ref.String(),
		HeadCommitId:        hc.Id.String(),
		HeadCommitUrl:       hc.Url.String(),
		HeadCommitAuthor:    hc.Author.Name.String(),
		HeadCommitCommitter: hc.Committer.Name.String(),
		HeadCommitMessage:   hc.Message.String(),
		HeadCommitTimestamp: hc.Timestamp.String(),
	}
	var emailBuf bytes.Buffer

	err := secretCommitEmailTmpl.Execute(&emailBuf, ctx)
	if err != nil {
		return err
	}

	if me.debug {
		log.Printf("Email message:\n%v\n", string(emailBuf.Bytes()))
	}
	return me.emailer.Send(me.fromAddr, me.recipients, emailBuf.Bytes())
}
