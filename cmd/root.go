package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	mode      string
	numAgents int
)

var rootCmd = &cobra.Command{
	Use:   "cde [name]",
	Short: "Create tmux coding environments",
	Args:  cobra.MaximumNArgs(1),
	RunE:  run,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&mode, "mode", "m", "ide", "session mode (ide, agent)")
	rootCmd.Flags().IntVarP(&numAgents, "num", "n", 2, "number of cc -w panes in agent mode")
}

func run(cmd *cobra.Command, args []string) error {
	name := ""
	if len(args) > 0 {
		name = args[0]
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		name = filepath.Base(cwd)
	}

	sessionName := strings.ReplaceAll(name, ".", "_") + "_" + mode

	switch mode {
	case "ide":
		return runIDE(sessionName)
	case "agent":
		return runAgent(sessionName, numAgents)
	default:
		return fmt.Errorf("unknown mode: %s (available: ide, agent)", mode)
	}
}

func tmux(args ...string) (string, error) {
	out, err := exec.Command("tmux", args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func tmuxOutput(args ...string) (string, error) {
	out, err := exec.Command("tmux", args...).Output()
	return strings.TrimSpace(string(out)), err
}

func hasSession(name string) bool {
	_, err := tmux("has-session", "-t", "="+name)
	return err == nil
}

func handleExistingSession(sessionName string) (bool, error) {
	if !hasSession(sessionName) {
		return false, nil
	}

	fmt.Printf("Session '%s' already exists\n", sessionName)
	fmt.Print("Open existing session? [Y/n] ")

	var reply string
	fmt.Scanln(&reply)
	reply = strings.TrimSpace(strings.ToLower(reply))

	if reply == "n" {
		_, err := tmux("kill-session", "-t", "="+sessionName)
		return false, err
	}

	return true, attachSession(sessionName)
}

func attachSession(sessionName string) error {
	fmt.Printf("\033]0;%s\007", sessionName)
	c := exec.Command("tmux", "attach-session", "-t", "="+sessionName)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func isGitRepo() bool {
	_, err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Output()
	return err == nil
}
