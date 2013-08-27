package hookworm

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"time"
)

type exitNoop struct{}

func (e *exitNoop) Error() string {
	return "exit noop 78"
}

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

func (sc *shellCommand) configure(config string) ([]byte, error) {
	if err := sc.preBuild(); err != nil {
		return []byte(""), err
	}

	return sc.runCmd(config, "configure")
}

func (sc *shellCommand) handleGithubPayload(payload string) ([]byte, error) {
	return sc.runCmd(payload, "handle", "github")
}

func (sc *shellCommand) handleTravisPayload(payload string) ([]byte, error) {
	return sc.runCmd(payload, "handle", "travis")
}

func (sc *shellCommand) preBuild() error {
	if sc.interpreter != "go run" {
		return nil
	}

	if err := sc.swapGoRunWithBinary(); err != nil {
		return err
	}

	return nil
}

func (sc *shellCommand) swapGoRunWithBinary() error {
	workingDir := os.Getenv("HOOKWORM_WORKING_DIR")
	if len(workingDir) < 1 {
		return fmt.Errorf("missing HOOKWORM_WORKING_DIR")
	}

	filePath := sc.filePath
	sc.filePath = ""
	sc.interpreter = "go"

	outfile := path.Join(workingDir, strings.Split(path.Base(filePath), ".")[0])
	_, err := sc.runCmd("", "build", "-o", outfile, filePath)
	if err != nil {
		return err
	}

	sc.interpreter = outfile

	return nil
}

func (sc *shellCommand) runCmd(stdin string, argv ...string) ([]byte, error) {
	var (
		cmd         *exec.Cmd
		commandArgs []string
		out         bytes.Buffer
	)

	if sc.filePath != "" {
		commandArgs = append(commandArgs, sc.filePath)
	}

	commandArgs = append(commandArgs, argv...)

	cmd = exec.Command(sc.interpreter, commandArgs...)
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return []byte(""), err
	}

	done := make(chan error)
	go func() { done <- cmd.Wait() }()

	select {
	case <-time.After(time.Duration(sc.timeout) * time.Second):
		err := cmd.Process.Kill()
		<-done
		return out.Bytes(), sc.errWrap(err)
	case err := <-done:
		return out.Bytes(), sc.errWrap(err)
	}

	panic("I should not be here")
}

func (sc *shellCommand) errWrap(err error) error {
	if err == nil {
		return nil
	}

	if msg, ok := err.(*exec.ExitError); ok {
		status := msg.Sys().(syscall.WaitStatus).ExitStatus()
		if status == 78 {
			err = &exitNoop{}
		}
	}

	return err
}
