package hookworm

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

	logger.Println("WARNING: no next handler?")
	return "", nil
}

func (th *topHandler) HandleTravisPayload(payload string) (string, error) {
	if th.next != nil {
		return th.next.HandleTravisPayload(payload)
	}

	logger.Println("WARNING: no next handler?")
	return "", nil
}

func (th *topHandler) NextHandler() Handler {
	return th.next
}

func (th *topHandler) SetNextHandler(nextHandler Handler) {
	th.next = nextHandler
}
