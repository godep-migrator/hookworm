package hookworm

import (
	"log"
)

type topHandler struct {
	next Handler
}

func newTopHandler() *topHandler {
	return &topHandler{}
}

func (me *topHandler) HandleGithubPayload(payload *GithubPayload) error {
	if me.next != nil {
		return me.next.HandleGithubPayload(payload)
	}

	log.Printf("WARNING: no next handler? %+v", me)
	return nil
}

func (me *topHandler) HandleTravisPayload(payload *TravisPayload) error {
	if me.next != nil {
		return me.next.HandleTravisPayload(payload)
	}

	log.Printf("WARNING: no next handler? %+v", me)
	return nil
}

func (me *topHandler) NextHandler() Handler {
	return me.next
}

func (me *topHandler) SetNextHandler(nextHandler Handler) {
	me.next = nextHandler
}
