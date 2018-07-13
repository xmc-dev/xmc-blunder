package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

const diffArgs = "-qBbEa"

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "Wrong number of arguments")
		os.Exit(1)
	}

	cmd := exec.Command("diff", diffArgs, os.Args[2], os.Args[3])
	err := cmd.Run()
	if err == nil {
		fmt.Fprintln(os.Stderr, "OK")
		fmt.Println("1.00")
	} else {
		e := err.(*exec.ExitError)
		sys := e.ProcessState.Sys().(syscall.WaitStatus)
		if sys.ExitStatus() > 1 {
			fmt.Fprintln(os.Stderr, "Error while diffing input and ok: ", err)
			os.Exit(1)
		} else {
			fmt.Fprintln(os.Stderr, "Incorrect")
			fmt.Println("0.00")
		}
	}
}
