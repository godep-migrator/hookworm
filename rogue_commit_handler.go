package hookworm

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
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

type RogueCommitHandler struct {
	debug                  bool
	emailer                *Emailer
	fromAddr               string
	recipients             []string
	policedBranches        []*regexp.Regexp
	policedBranchesStrings []string
	nextHandler            Handler
}

type rogueCommitEmailContext struct {
	From                  string
	Recipients            string
	Date                  string
	MessageId             string
	Hostname              string
	Repo                  string
	Ref                   string
	PolicedBranches       string
	RepoUrl               string
	HeadCommitId          string
	HeadCommitUrl         string
	HeadCommitAuthor      string
	HeadCommitCommitter   string
	HeadCommitMessageText string
	HeadCommitMessageHtml string
	HeadCommitTimestamp   string
}

func NewRogueCommitHandler(cfg *HandlerConfig) *RogueCommitHandler {
	handler := &RogueCommitHandler{
		debug:           cfg.Debug,
		emailer:         NewEmailer(cfg.EmailUri),
		fromAddr:        cfg.EmailFromAddr,
		recipients:      cfg.EmailRcpts,
		policedBranches: strsToRegexes(cfg.PolicedBranches),
	}

	for _, re := range handler.policedBranches {
		handler.policedBranchesStrings = append(handler.policedBranchesStrings, re.String())
	}

	return handler
}

func (me *RogueCommitHandler) HandlePayload(payload *Payload) error {
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

func (me *RogueCommitHandler) SetNextHandler(handler Handler) {
	me.nextHandler = handler
}

func (me *RogueCommitHandler) NextHandler() Handler {
	return me.nextHandler
}

func (me *RogueCommitHandler) isPolicedBranch(ref string) bool {
	sansRefsHeads := strings.Replace(ref, "refs/heads/", "", 1)
	for _, ref := range me.policedBranches {
		if ref.MatchString(sansRefsHeads) {
			return true
		}
	}
	return false
}

func (me *RogueCommitHandler) alert(payload *Payload) error {
	log.Printf("WARNING rogue commit! %+v, head commit: %+v\n",
		payload, payload.HeadCommit)
	if len(me.recipients) == 0 {
		log.Println("No email recipients specified, so no emailing!")
		return nil
	}

	hc := payload.HeadCommit
	ctx := &rogueCommitEmailContext{
		From:       me.fromAddr,
		Recipients: strings.Join(me.recipients, ", "),
		Date:       time.Now().UTC().Format(rfc2822DateFmt),
		MessageId:  fmt.Sprintf("%v", time.Now().UTC().UnixNano()),
		Hostname:   hostname,
		Repo: fmt.Sprintf("%s/%s", payload.Repository.Owner.Name.String(),
			payload.Repository.Name.String()),
		Ref:                   payload.Ref.String(),
		PolicedBranches:       strings.Join(me.policedBranchesStrings, ", "),
		RepoUrl:               payload.Repository.Url.String(),
		HeadCommitId:          hc.Id.String(),
		HeadCommitUrl:         hc.Url.String(),
		HeadCommitAuthor:      hc.Author.Name.String(),
		HeadCommitCommitter:   hc.Committer.Name.String(),
		HeadCommitMessageText: hc.Message.String(),
		HeadCommitMessageHtml: hc.Message.Html(),
		HeadCommitTimestamp:   hc.Timestamp.String(),
	}
	var emailBuf bytes.Buffer

	err := rogueCommitEmailTmpl.Execute(&emailBuf, ctx)
	if err != nil {
		return err
	}

	if me.debug {
		log.Printf("Email message:\n%v\n", string(emailBuf.Bytes()))
	}
	return me.emailer.Send(me.fromAddr, me.recipients, emailBuf.Bytes())
}
