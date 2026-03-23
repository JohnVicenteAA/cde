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
	rootCmd.Flags().StringVarP(&mode, "mode", "m", "ide", "session mode (ide, wtree, mrepo)")
	rootCmd.Flags().IntVarP(&numAgents, "num", "n", 2, "number of claude --worktree panes in wtree mode")
}

func sessionName(name, mode string) string {
	return strings.ReplaceAll(name, ".", "_") + "_" + mode
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

	sn := sessionName(name, mode)

	windowTitle := name + ": " + mode

	switch mode {
	case "ide":
		return runIDE(sn, windowTitle)
	case "wtree":
		return runWtree(sn, numAgents, windowTitle)
	case "mrepo":
		return runMrepo(sn, windowTitle)
	default:
		return fmt.Errorf("unknown mode: %s (available: ide, wtree, mrepo)", mode)
	}
}

func handleExistingSession(sessionName string) (bool, error) {
	if !runner.HasSession(sessionName) {
		return false, nil
	}

	fmt.Printf("Session '%s' already exists\n", sessionName)
	fmt.Print("Open existing session? [Y/n] ")

	var reply string
	fmt.Scanln(&reply)
	reply = strings.TrimSpace(strings.ToLower(reply))

	if reply == "n" {
		_, err := runner.Run("kill-session", "-t", "="+sessionName)
		return false, err
	}

	return true, runner.Attach(sessionName)
}

var isGitRepo = func() bool {
	_, err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Output()
	return err == nil
}
