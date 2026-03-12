package cmd

import (
	"os"
	"os/exec"
	"strings"
)

type TmuxRunner interface {
	Run(args ...string) (string, error)
	Attach(sessionName string) error
	HasSession(name string) bool
}

type realTmux struct{}

func (r *realTmux) Run(args ...string) (string, error) {
	out, err := exec.Command("tmux", args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (r *realTmux) Attach(sessionName string) error {
	c := exec.Command("tmux", "attach-session", "-t", "="+sessionName)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func (r *realTmux) HasSession(name string) bool {
	_, err := r.Run("has-session", "-t", "="+name)
	return err == nil
}

var runner TmuxRunner = &realTmux{}
