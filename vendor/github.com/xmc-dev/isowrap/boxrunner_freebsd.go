package isowrap

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// BoxRunner is a Runner based on jail
type BoxRunner struct {
	B *Box
}

func (br *BoxRunner) rctl(flag, rule string) error {
	params := []string{}
	params = append(
		params,
		flag,
		fmt.Sprintf("jail:isowrap%d:%s/jail", br.B.ID, rule),
	)
	_, _, _, err := Exec("rctl", params...)
	return err
}

// Init creates the jail and sets the rctl's
func (br *BoxRunner) Init() error {
	p := filepath.Join(os.TempDir(), fmt.Sprintf("isowrap%d", br.B.ID))
	err := os.MkdirAll(filepath.Join(p, "root"), os.ModePerm)
	if err != nil {
		return err
	}

	// Create jail
	params := []string{}
	params = append(
		params,
		"-c",
		fmt.Sprintf("name=isowrap%d", br.B.ID),
		"path="+p,
		"persist",
	)
	_, _, _, err = Exec("jail", params...)
	if err != nil {
		return err
	}

	cl := func(rule string) {
		if err != nil {
			return
		}
		err = br.rctl("-a", rule)
	}
	// Set resource limits
	if br.B.Config.MemoryLimit > 0 {
		cl(fmt.Sprintf("memoryuse:sigsegv=%dK", br.B.Config.MemoryLimit))
	}
	if br.B.Config.StackLimit > 0 {
		cl(fmt.Sprintf("stacksize:sigsegv=%dK", br.B.Config.MemoryLimit))
	}
	if br.B.Config.MaxProc > 0 {
		cl(fmt.Sprintf("maxproc:deny=%d", br.B.Config.MaxProc))
	}
	if err != nil {
		return err
	}

	br.B.Path = p
	return nil
}

// Run executes jexec to execute the given command.
func (br *BoxRunner) Run(command string, args ...string) (result RunResult, err error) {
	var bout, berr bytes.Buffer

	params := []string{}
	params = append(
		params,
		fmt.Sprintf("isowrap%d", br.B.ID),
		"/"+command,
	)
	params = append(params, args...)

	result.ErrorType = NoError

	cmd := exec.Command("jexec", params...)
	cmd.Stdout = &bout
	cmd.Stderr = &berr
	cmd.Env = []string{}
	for _, e := range br.B.Config.Env {
		// If no value given, inherit environment variable from the system
		if e.Value == "" {
			cmd.Env = append(cmd.Env, e.Var+"="+os.Getenv(e.Var))
		} else {
			cmd.Env = append(cmd.Env, e.Var+"="+e.Value)
		}
	}

	startTime := time.Now()
	err = cmd.Start()
	if err != nil {
		return
	}
	proc := cmd.Process
	isDone := false

	done := make(chan error, 1)
	waitForProc := func() {
		done <- cmd.Wait()
	}
	doneProcess := func() {
		if _, ok := err.(*exec.ExitError); !ok {
			return
		} else {
			err = nil
		}
		isDone = true
	}

	if br.B.Config.WallTime > 0 {
		go waitForProc()

		select {
		case <-time.After(br.B.Config.WallTime):
			if !isDone {
				if err = proc.Kill(); err != nil {
					return
				}
				result.ErrorType = Timeout
			}
		case err = <-done:
			doneProcess()
		}
	} else {
		waitForProc()
		doneProcess()
	}
	cmd.Wait()
	wallTime := time.Since(startTime)
	result.WallTime = wallTime

	result.Stdout = string(bout.Bytes())
	result.Stderr = string(berr.Bytes())

	if result.ErrorType != NoError {
		return
	}

	state := cmd.ProcessState

	// state is nil if the process was killed
	if state != nil {
		ws, ok := state.Sys().(syscall.WaitStatus)
		if !ok {
			result.ErrorType = InternalError
			return
		}

		us, ok := state.SysUsage().(*syscall.Rusage)
		if !ok {
			result.ErrorType = InternalError
			return
		}
		result.ExitCode = ws.ExitStatus()
		// Return code by convention
		if ws.Signaled() {
			result.ExitCode = 128 + int(ws.Signal())
			result.ErrorType = KilledBySignal
			result.Signal = syscall.Signal(ws.Signal())
		}
		result.CPUTime = state.SystemTime() + state.UserTime()
		result.MemUsed = uint(us.Maxrss)
	}

	if result.ExitCode != 0 && result.ErrorType == NoError {
		result.ErrorType = RunTimeError
	}

	return
}

// Cleanup deletes the jail and its directory.
func (br *BoxRunner) Cleanup() error {
	// stop jail
	params := []string{}
	params = append(
		params,
		"-r",
		fmt.Sprintf("isowrap%d", br.B.ID),
	)
	_, _, _, err := Exec("rctl", "-r", fmt.Sprintf("jail:isowrap%d", br.B.ID))
	if err != nil {
		return err
	}
	_, _, _, err = Exec("jail", params...)
	if err != nil {
		return err
	}
	err = os.RemoveAll(br.B.Path)
	if err != nil {
		return err
	}
	return nil
}
