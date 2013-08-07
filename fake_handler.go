package hookworm

import (
	"log"
)

type fakeHandler struct {
	next Handler
}

func newFakeHandler() *fakeHandler {
	return &fakeHandler{}
}

func (me *fakeHandler) HandleGithubPayload(payload *GithubPayload) error {
	if me.next != nil {
		return me.next.HandleGithubPayload(payload)
	}

	log.Printf("WARNING: no next handler? %+v", me)
	return nil
}

func (me *fakeHandler) NextHandler() Handler {
	return me.next
}

func (me *fakeHandler) SetNextHandler(nextHandler Handler) {
	me.next = nextHandler
}
