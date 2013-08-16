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

func (me *topHandler) HandleGithubPayload(payload string) error {
	if me.next != nil {
		return me.next.HandleGithubPayload(payload)
	}

	log.Println("WARNING: no next handler?")
	return nil
}

func (me *topHandler) HandleTravisPayload(payload string) error {
	if me.next != nil {
		return me.next.HandleTravisPayload(payload)
	}

	log.Println("WARNING: no next handler?")
	return nil
}

func (me *topHandler) NextHandler() Handler {
	return me.next
}

func (me *topHandler) SetNextHandler(nextHandler Handler) {
	me.next = nextHandler
}
