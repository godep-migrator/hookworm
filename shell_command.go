package hookworm

import (
	"fmt"
	"os"
	"os/exec"
	"path"
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
	if err := me.preBuild(); err != nil {
		return err
	}

	return me.runCmd(config, "configure")
}

func (me *shellCommand) handleGithubPayload(payload string) error {
	return me.runCmd(payload, "handle", "github")
}

func (me *shellCommand) handleTravisPayload(payload string) error {
	return me.runCmd(payload, "handle", "travis")
}

func (me *shellCommand) preBuild() error {
	if me.interpreter != "go run" {
		return nil
	}

	if err := me.swapGoRunWithBinary(); err != nil {
		return err
	}

	return nil
}

func (me *shellCommand) swapGoRunWithBinary() error {
	workingDir := os.Getenv("HOOKWORM_WORKING_DIR")
	if len(workingDir) < 1 {
		return fmt.Errorf("Missing HOOKWORM_WORKING_DIR!")
	}

	filePath := me.filePath
	me.filePath = ""
	me.interpreter = "go"

	outfile := path.Join(workingDir, strings.Split(path.Base(filePath), ".")[0])
	err := me.runCmd("", "build", "-o", outfile, filePath)
	if err != nil {
		return err
	}

	me.interpreter = outfile

	return nil
}

func (me *shellCommand) runCmd(stdin string, argv ...string) error {
	var (
		cmd         *exec.Cmd
		commandArgs []string
	)

	if me.filePath != "" {
		commandArgs = append(commandArgs, me.filePath)
	}

	commandArgs = append(commandArgs, argv...)

	cmd = exec.Command(me.interpreter, commandArgs...)
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
