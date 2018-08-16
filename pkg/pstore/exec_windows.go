package pstore

import (
	"os/exec"
	"os"
	"fmt"
)

func ExecCommand(args []string) {
	if len(args) == 0 {
		abort(usageError, "no command specified")
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()

	if err != nil {
		fmt.Println("err: %s", err.Error())
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
