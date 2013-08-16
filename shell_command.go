package hookworm

import (
	"os"
	"os/exec"
	"strings"
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

func (me *shellCommand) configure(config string) error {
	return me.runCmd(config, "configure")
}

func (me *shellCommand) handleGithubPayload(payload string) error {
	return me.runCmd(payload, "handle", "github")
}

func (me *shellCommand) handleTravisPayload(payload string) error {
	return me.runCmd(payload, "handle", "travis")
}

func (me *shellCommand) runCmd(stdin string, argv ...string) error {
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
	cmd.Stdin = strings.NewReader(stdin)
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
