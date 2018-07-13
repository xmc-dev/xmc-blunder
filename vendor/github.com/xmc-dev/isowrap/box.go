package isowrap

import (
	"bytes"
	"io"
	"os"
	"time"
)

// EnvPair represents an environment variable made of a key and a value.
type EnvPair struct {
	Var   string
	Value string
}

// BoxError represents an error encountered after running the program in the box.
type BoxError int

// RunResult represents the result of running the prograim in the box.
type RunResult struct {
	ExitCode int

	CPUTime   time.Duration
	WallTime  time.Duration
	MemUsed   uint
	Signal    os.Signal
	ErrorType BoxError
}

// BoxConfig contains configuration data for the BoxRunner
type BoxConfig struct {
	CPUTime      time.Duration
	WallTime     time.Duration
	MemoryLimit  uint
	StackLimit   uint
	MaxProc      uint
	ShareNetwork bool
	Env          []EnvPair
}

// Runner is an interface for various program isolating methods
type Runner interface {
	Init() error
	Run(stdin io.Reader, stdout, stderr io.Writer, command string, args ...string) (RunResult, error)
	Cleanup() error
}

const (
	// NoError means that no error has been returned by the box runner
	NoError BoxError = iota

	// RunTimeError means that an error was raised at run time. Probably non-zero status.
	RunTimeError = iota

	// KilledBySignal means that the program was killed after getting a signal.
	// Probably because of resource error or memory violations.
	KilledBySignal = iota

	// Timeout means that the running program exceeded the target timeout.
	Timeout = iota

	// InternalError means that the Runner encountered an error.
	InternalError = iota

	// MemoryExceeded means that the process tried to use more memory than provided
	MemoryExceeded = iota
)

func (be BoxError) String() string {
	var s string
	switch be {
	case NoError:
		s = "NoError"
	case RunTimeError:
		s = "RunTimeError"
	case KilledBySignal:
		s = "KilledBySignal"
	case Timeout:
		s = "Timeout"
	case InternalError:
		s = "InternalError"
	case MemoryExceeded:
		s = "MemoryExceeded"
	}

	return s
}

// Box represents an isolated environment
type Box struct {
	Config BoxConfig
	Path   string
	ID     uint

	runner Runner
}

// DefaultBoxConfig returns a new instance of the default box config
func DefaultBoxConfig() BoxConfig {
	bc := BoxConfig{}

	bc.Env = make([]EnvPair, 1)
	bc.Env[0].Var = "LIBC_FATAL_STDERR_"
	bc.Env[0].Value = "1"

	return bc
}

// NewBox returns a new Box instance
func NewBox() *Box {
	b := Box{}
	b.Config = DefaultBoxConfig()
	b.runner = &BoxRunner{&b}

	return &b
}

// Init calls the runner's Init function.
func (b *Box) Init() error {
	return b.runner.Init()
}

// Run calls the runner's Run function
func (b *Box) Run(stdin io.Reader, stdout, stderr io.Writer, command string, args ...string) (RunResult, error) {
	return b.runner.Run(stdin, stdout, stderr, command, args...)
}

func (b *Box) RunOutput(command string, args ...string) (string, string, RunResult, error) {
	var stdout, stderr bytes.Buffer
	result, err := b.Run(os.Stdin, &stdout, &stderr, command, args...)
	return stdout.String(), stderr.String(), result, err
}

// Cleanup calls the runner's Cleanup function.
func (b *Box) Cleanup() error {
	return b.runner.Cleanup()
}
