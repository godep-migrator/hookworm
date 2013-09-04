package hookworm

import (
	"encoding/json"
	"log"
	"path"
)

type shellHandler struct {
	command shellCommand
	cfg     *HandlerConfig
	next    Handler
}

var (
	interpreterMap = map[string]string{
		".js": "node",
		".py": "python",
		".pl": "perl",
		".rb": "ruby",
		".sh": "bash",
	}
)

func newShellHandler(filePath string, cfg *HandlerConfig) (*shellHandler, error) {
	handler := &shellHandler{}

	fileExtention := path.Ext(filePath)

	handler.cfg = cfg

	if interpreter, ok := interpreterMap[fileExtention]; ok {
		handler.command = newShellCommand(interpreter, filePath, cfg.WormTimeout)
	}

	if err := handler.configure(); err != nil {
		return nil, err
	}

	return handler, nil
}

func (sh *shellHandler) configure() error {
	configJSON, err := json.Marshal(sh.cfg)
	if err != nil {
		log.Printf("Error JSON-marshalling config: %v", err)
	}

	_, err = sh.command.configure(string(configJSON))
	return err
}

func (sh *shellHandler) HandleGithubPayload(payload string) (string, error) {
	if sh.cfg.Debug {
		log.Printf("Sending github payload to %+v\n", sh)
	}

	noop := false
	outBytes, err := sh.command.handleGithubPayload(payload)
	out := string(outBytes)

	if _, noop = err.(*exitNoop); noop {
		out = payload
	}

	if err != nil && !noop {
		return out, err
	}

	if sh.next != nil {
		return sh.next.HandleGithubPayload(out)
	}

	return out, nil
}

func (sh *shellHandler) HandleTravisPayload(payload string) (string, error) {
	if sh.cfg.Debug {
		log.Printf("Sending travis payload to %+v\n", sh)
	}
	noop := false
	outBytes, err := sh.command.handleTravisPayload(payload)
	out := string(outBytes)

	if _, noop = err.(*exitNoop); noop {
		out = payload
	}

	if err != nil && !noop {
		return out, err
	}

	if sh.next != nil {
		return sh.next.HandleTravisPayload(out)
	}

	return out, nil
}

func (sh *shellHandler) SetNextHandler(n Handler) {
	sh.next = n
}

func (sh *shellHandler) NextHandler() Handler {
	return sh.next
}
