package common

import (
	"os"
	"os/exec"
)

type Program struct {
	Source     string
	Executable string
	Language   Language
	Manager    ProgramManager
}

type ProgramManager interface {
	Version() *exec.Cmd
	Compile(p *Program, args ...string) *exec.Cmd
	Execute(p *Program, args ...string) *exec.Cmd
}

func (p *Program) Version() *exec.Cmd {
	return p.Manager.Version()
}

func (p *Program) Compile(args ...string) *exec.Cmd {
	return p.Manager.Compile(p, args...)
}

func (p *Program) Execute(args ...string) *exec.Cmd {
	return p.Manager.Execute(p, args...)
}

func NewProgram(source, executable string, language Language) *Program {
	var manager ProgramManager
	switch language {
	case LangC:
		manager = &CManager{}
	case LangCPP:
		manager = &CPPManager{}
	case LangGo:
		manager = &GoManager{}
	}

	return &Program{
		Source:     source,
		Executable: executable,
		Language:   language,
		Manager:    manager,
	}
}

type CManager struct{}

func (*CManager) Version() *exec.Cmd {
	return exec.Command("clang", "--version")
}

func (*CManager) Compile(p *Program, args ...string) *exec.Cmd {
	args = append([]string{"-o", p.Executable, "-DONLINE_JUDGE", "-Wall", "-O2", "-static", "-std=c11", "-lm", p.Source}, args...)
	return exec.Command("clang", args...)
}

func (*CManager) Execute(p *Program, args ...string) *exec.Cmd {
	return exec.Command(p.Executable, args...)
}

type CPPManager struct{}

func (*CPPManager) Version() *exec.Cmd {
	return exec.Command("clang++", "--version")
}

func (*CPPManager) Compile(p *Program, args ...string) *exec.Cmd {
	args = append([]string{"-o", p.Executable, "-DONLINE_JUDGE", "-Wall", "-O2", "-static", "-std=c++11", "-lm", p.Source}, args...)
	return exec.Command("clang++", args...)
}

func (*CPPManager) Execute(p *Program, args ...string) *exec.Cmd {
	return exec.Command(p.Executable, args...)
}

type GoManager struct{}

func (*GoManager) Version() *exec.Cmd {
	return exec.Command("go", "version")
}

func (*GoManager) Compile(p *Program, args ...string) *exec.Cmd {
	args = append([]string{"build", "-a", "-installsuffix", "cgo", "-ldflags", "-s", "-o", p.Executable, p.Source}, args...)
	cmd := exec.Command("go", args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "CGO_ENABLED=0")
	return cmd
}

func (*GoManager) Execute(p *Program, args ...string) *exec.Cmd {
	return exec.Command(p.Executable, args...)
}
