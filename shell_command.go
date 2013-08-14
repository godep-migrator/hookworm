package hookworm

import (
	"bytes"
	"os"
	"os/exec"
	"time"
)

type shellCommand struct {
	interpreter string
	filePath    string
	timeout     int
}

func newShellCommand(interpreter, filePath string, timeout int) shellCommand {
	return shellCommand{
		interpreter: interpreter,
		filePath:    filePath,
		timeout:     timeout,
	}
}

func (me *shellCommand) configure(configJSON []byte) error {
	return me.runCmd(configJSON, "configure")
}

func (me *shellCommand) handleGithubPayload(payloadJSON []byte) error {
	return me.runCmd(payloadJSON, "handle", "github")
}

func (me *shellCommand) handleTravisPayload(payloadJSON []byte) error {
	return me.runCmd(payloadJSON, "handle", "travis")
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

	err := cmd.Start()
	if err != nil {
		return err
	}

	done := make(chan error)
	go func() { done <- cmd.Wait() }()

	select {
	case <-time.After(time.Duration(me.timeout) * time.Second):
		err := cmd.Process.Kill()
		<-done
		return err
	case err := <-done:
		return err
	}

	panic("I should not be here")
}
