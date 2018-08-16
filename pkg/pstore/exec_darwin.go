package pstore

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func ExecCommand(args []string) {
	if len(args) == 0 {
		abort(usageError, "no command specified")
	}
	commandName := args[0]
	commandPath, err := exec.LookPath(commandName)
	if err != nil {
		abort(commandNotFoundError, fmt.Sprintf("cannot find '%s'", commandName))
	}
	err = syscall.Exec(commandPath, args, os.Environ())
	if err != nil {
		abort(execError, err)
	}
}
