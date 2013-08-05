package hookworm

import (
	"bytes"
	"os"
	"os/exec"
)

type shellCommand struct {
	interpreter string
	filePath    string
}

func newShellCommand(interpreter, filePath string) shellCommand {
	return shellCommand{
		interpreter: interpreter,
		filePath:    filePath,
	}
}

func (me *shellCommand) configure(configJSON []byte) error {
	return me.runCmd(configJSON, "configure")
}

func (me *shellCommand) handleGithubPayload(payloadJSON []byte) error {
	return me.runCmd(payloadJSON, "handle", "github")
}

func (me *shellCommand) runCmd(stdin []byte, argv ...string) error {
	var cmd *exec.Cmd

	var (
		interpreter string
		commandArgs []string
	)

	if me.interpreter == "go run" {
		interpreter = "go"
		commandArgs = append(commandArgs, "run", me.filePath)
	} else {
		interpreter = me.interpreter
		commandArgs = append(commandArgs, me.filePath)
	}

	commandArgs = append(commandArgs, argv...)

	cmd = exec.Command(interpreter, commandArgs...)
	cmd.Stdin = bytes.NewReader(stdin)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
