package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Cmd struct {
	cmds []*exec.Cmd
}

func NewCmd(c *exec.Cmd) *Cmd {
	cmd := &Cmd{}
	if c != nil {
		cmd.cmds = append(cmd.cmds, c)
	}
	return cmd
}

func (cmd *Cmd) AddCmd(c *exec.Cmd) {
	cmd.cmds = append(cmd.cmds, c)
}

func (cmd *Cmd) Run() error {
	je := &joinError{}
	for _, c := range cmd.cmds {
		printlnFunc("Running:", strings.Join(c.Args, " "))
		if err := c.Run(); err != nil {
			je.errs = append(je.errs, err)
		}
	}
	if len(je.errs) == 0 {
		return nil
	}
	return je
}

func clangFormat(files []string) *Cmd {
	return NewCmd(makeCmd("clang-format", append([]string{"-i"}, files...)...))
}

func goFmt(files []string) *Cmd {
	dirs := map[string][]string{}
	for _, file := range files {
		dir := filepath.Dir(file)
		dirs[dir] = append(dirs[dir], file)
	}
	cmd := NewCmd(nil)
	for _, files := range dirs {
		cmd.AddCmd(makeCmd("go", append([]string{"fmt"}, files...)...))
	}
	return cmd
}

func rustFmt(files []string) *Cmd {
	return NewCmd(makeCmd(
		"rustfmt",
		append([]string{"--edition=2021"}, files...)...,
	))
}

func makeCmd(prog string, args ...string) *exec.Cmd {
	cmd := exec.Command(prog, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd
}

type joinError struct {
	errs []error
}

func (je *joinError) Error() string {
	b := []byte(je.errs[0].Error())
	for _, err := range je.errs[1:] {
		b = append(b, '\n')
		b = append(b, err.Error()...)
	}
	return string(b)
}
