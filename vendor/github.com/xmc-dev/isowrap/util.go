package isowrap

import (
	"io"
	"os"
	"os/exec"
	"time"
)

// ExecResult holds information about the program after execution.
type ExecResult struct {
	State    *os.ProcessState
	WallTime time.Duration
}

// Exec executes a command and returns its stdout, stderr and exit status
func Exec(stdin io.Reader, stdout, stderr io.Writer, program string, args ...string) (result ExecResult, err error) {
	cmd := exec.Command(program, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	start := time.Now()

	err = cmd.Run()
	elapsed := time.Since(start)
	if err != nil {
		return
	}
	result.State = cmd.ProcessState
	result.WallTime = elapsed
	return
}
