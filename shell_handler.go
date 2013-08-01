package hookworm

import (
	"encoding/json"
	"log"
	"strings"
)

type ShellHandler struct {
	command    Command
	config     *HandlerConfig
	ConfigJSON string
	next       Handler
}

var (
	ExtensionToInterpretter = make(map[string]string)
)

func init() {
	ExtensionToInterpretter[".rb"] = "ruby"
	ExtensionToInterpretter[".py"] = "python"
	ExtensionToInterpretter[".sh"] = "bash"
	ExtensionToInterpretter[".go"] = "go run"
}

// assumed to be executable at this point
func NewShellHandler(filePath string, cfg *HandlerConfig) *ShellHandler {
	result := &ShellHandler{}

	fileExtention := getFileExtension(filePath)

	result.config = cfg
	result.configureSelf()

	s := ExtensionToInterpretter[fileExtention]
	result.command = NewCommand(s, filePath)

	return result
}

func (me ShellHandler) configureSelf() {

	marshalled_json, err := json.Marshal(me.config)
	if err != nil {
		log.Println("error in marshahlling the json")
	}
	me.ConfigJSON = string(marshalled_json)
}

func (me ShellHandler) Execute() {
	me.command.RunAndWait()
}

func (me ShellHandler) HandleGithubPayload(payload *GithubPayload) error {
	me.Execute()
	return nil
}

func (me ShellHandler) SetNextHandler(n Handler) {
	me.next = n
}

func (me ShellHandler) NextHandler() Handler {
	return me.next
}

func getFileExtension(filePath string) string {
	var result string

	index := strings.LastIndex(filePath, ".")
	result = filePath[index:]

	return result

}
