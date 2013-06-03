package hookworm

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"strings"
	"text/template"
	"time"
)

var (
	rfc2822DateFmt        = "Mon, 01 Jan 15:04:05 -0700"
	secretCommitEmailTmpl = template.Must(template.New("email").Parse(`From: {{.From}}
To: {{.Recipients}}
Subject: [hookworm] Secret commit! {{.Repo}} {{.Ref}} {{.HeadCommitId}}
Date: {{.Date}}
Message-ID: {{.MessageId}}
Content-Type: text/html; charset=utf8

<h1>Secret commit detected on {{.Repo}} {{.Ref}}</h1>

<dl>
  <dt>Id</dt><dd><a href="{{.HeadCommitUrl}}">{{.HeadCommitId}}</a></dd>
  <dt>Message</dt><dd>{{.HeadCommitMessage}}</dd>
  <dt>Author</dt><dd>{{.HeadCommitAuthor}}</dd>
  <dt>Committer</dt><dd>{{.HeadCommitCommitter}}</dd>
  <dt>Timestamp</dt><dd>{{.HeadCommitTimestamp}}</dd>
</dl>
`))
)

type SecretSquirrelCommitHandler struct {
	emailer        *Emailer
	fromAddr       string
	recipients     []string
	stableBranches []string
	nextHandler    Handler
}

type secretCommitEmailContext struct {
	From                string
	Recipients          string
	Date                string
	MessageId           string
	Repo                string
	Ref                 string
	HeadCommitId        string
	HeadCommitUrl       string
	HeadCommitAuthor    string
	HeadCommitCommitter string
	HeadCommitMessage   string
	HeadCommitTimestamp string
}

func (me *SecretSquirrelCommitHandler) HandlePayload(payload *Payload) error {
	me.checkIfSecretSquirrelCommit(payload)
	return nil
}

func (me *SecretSquirrelCommitHandler) SetNextHandler(handler Handler) {
	me.nextHandler = handler
}

func (me *SecretSquirrelCommitHandler) NextHandler() Handler {
	return me.nextHandler
}

func (me *SecretSquirrelCommitHandler) checkIfSecretSquirrelCommit(payload *Payload) {
	if !me.isStableBranch(payload.Ref.String()) {
		return
	}

	if payload.IsPullRequestMerge() {
		return
	}

	if err := me.alert(payload); err != nil {
		log.Printf("ERROR sending alert: %+v\n", err)
		return
	}

	log.Printf("Sent alert to %+v\n", me.recipients)
}

func (me *SecretSquirrelCommitHandler) isStableBranch(ref string) bool {
	return sort.SearchStrings(me.stableBranches, ref) > -1
}

func (me *SecretSquirrelCommitHandler) alert(payload *Payload) error {
	// FIXME use the emailer thing here
	log.Printf("WARNING secret squirrel commit! %+v\n", payload)
	if len(me.recipients) == 0 {
		log.Println("No email recipients specified, so no emailing!")
	}

	hc := payload.HeadCommit
	ctx := &secretCommitEmailContext{
		From:       me.fromAddr,
		Recipients: strings.Join(me.recipients, ", "),
		Date:       time.Now().UTC().Format(rfc2822DateFmt),
		MessageId:  fmt.Sprintf("%v", time.Now().UTC().UnixNano()),
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

	log.Printf("Email message:\n%v\n", string(emailBuf.Bytes()))
	return me.emailer.Send(me.fromAddr, me.recipients, emailBuf.Bytes())
}
