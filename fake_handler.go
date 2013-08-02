package hookworm

type FakeHandler struct {
	next Handler
}

func NewFakeHandler() FakeHandler { return FakeHandler{} }

func (me FakeHandler) HandleGithubPayload(payload *GithubPayload) error {
	return nil
}

func (me FakeHandler) NextHandler() Handler {
	return me.next
}

func (me FakeHandler) SetNextHandler(nextHandler Handler) {
	me.next = nextHandler
}
