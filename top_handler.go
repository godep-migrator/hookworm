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

func (th *topHandler) HandleGithubPayload(payload string) (string, error) {
	if th.next != nil {
		return th.next.HandleGithubPayload(payload)
	}

	log.Println("WARNING: no next handler?")
	return "", nil
}

func (th *topHandler) HandleTravisPayload(payload string) (string, error) {
	if th.next != nil {
		return th.next.HandleTravisPayload(payload)
	}

	log.Println("WARNING: no next handler?")
	return "", nil
}

func (th *topHandler) NextHandler() Handler {
	return th.next
}

func (th *topHandler) SetNextHandler(nextHandler Handler) {
	th.next = nextHandler
}
