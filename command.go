package hookworm

import (
	"os/exec"
)

type Command struct {
	Interpretter string
	FilePath     string
}

func NewCommand(interpretter string, filePath string) Command {
	return Command{
		Interpretter: interpretter,
		FilePath:     filePath,
	}
}

func (me Command) RunAndWait() int {

	var cmd *exec.Cmd
	var err error

	if me.Interpretter == "go run" {
		cmd = exec.Command("go", "run", me.FilePath)
	} else {
		cmd = exec.Command(me.Interpretter, me.FilePath)
	}
	err = cmd.Run()

	if err != nil {
		return -1
	}

	return 0
}
