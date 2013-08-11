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
		".go": "go run",
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

func (me *shellHandler) configure() error {
	configJSON, err := json.Marshal(me.cfg)
	if err != nil {
		log.Printf("Error JSON-marshalling config: %v", err)
	}

	return me.command.configure(configJSON)
}

func (me *shellHandler) HandleGithubPayload(payload *GithubPayload) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if me.cfg.Debug {
		log.Printf("Sending github payload to %+v\n", me)
	}
	return me.command.handleGithubPayload(payloadJSON)
}

func (me *shellHandler) SetNextHandler(n Handler) {
	me.next = n
}

func (me *shellHandler) NextHandler() Handler {
	return me.next
}
