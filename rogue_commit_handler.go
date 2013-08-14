package hookworm

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

type RogueCommitHandler struct {
	debug                  bool
	emailer                *Emailer
	fromAddr               string
	recipients             []string
	watchedBranches        []*regexp.Regexp
	watchedPaths           []*regexp.Regexp
	watchedBranchesStrings []string
	watchedPathsStrings    []string
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
	WatchedBranches       []string
	WatchedPaths          []string
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
		watchedBranches: strsToRegexes(cfg.WatchedBranches),
		watchedPaths:    strsToRegexes(cfg.WatchedPaths),
	}

	for _, re := range handler.watchedBranches {
		handler.watchedBranchesStrings = append(handler.watchedBranchesStrings, re.String())
	}

	for _, re := range handler.watchedPaths {
		handler.watchedPathsStrings = append(handler.watchedPathsStrings, re.String())
	}

	return handler
}

func (me *RogueCommitHandler) HandleGithubPayload(payload *GithubPayload) error {
	if !me.isWatchedBranch(payload.Ref.String()) {
		if me.debug {
			log.Printf("%v is not a watched branch, yay!\n", payload.Ref.String())
		}
		return nil
	}

	if me.debug {
		log.Printf("%v is a watched branch!\n", payload.Ref.String())
	}

	hcId := payload.HeadCommit.Id.String()

	if payload.IsPullRequestMerge() {
		if me.debug {
			log.Printf("%v is a pull request merge, yay!\n", hcId)
		}
		return nil
	}

	if me.debug {
		log.Printf("%v is not a pull request merge!\n", hcId)
	}

	if !me.hasWatchedPath(payload.Paths()) {
		if me.debug {
			log.Printf("%v does not contain watched paths, yay!\n", hcId)
		}
		return nil
	}

	if me.debug {
		log.Printf("%v contains watched paths!\n", hcId)
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

func (me *RogueCommitHandler) HandleTravisPayload(*TravisPayload) error {
	return nil
}

func (me *RogueCommitHandler) SetNextHandler(handler Handler) {
	me.nextHandler = handler
}

func (me *RogueCommitHandler) NextHandler() Handler {
	return me.nextHandler
}

func (me *RogueCommitHandler) isWatchedBranch(ref string) bool {
	sansRefsHeads := strings.Replace(ref, "refs/heads/", "", 1)
	for _, branchRe := range me.watchedBranches {
		if branchRe.MatchString(sansRefsHeads) {
			return true
		}
	}
	return false
}

func (me *RogueCommitHandler) hasWatchedPath(paths []string) bool {
	for _, pathRe := range me.watchedPaths {
		for _, path := range paths {
			if pathRe.MatchString(path) {
				return true
			}
		}
	}
	return false
}

func (me *RogueCommitHandler) alert(payload *GithubPayload) error {
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
		WatchedBranches:       me.watchedBranchesStrings,
		WatchedPaths:          me.watchedPathsStrings,
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
