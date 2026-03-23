package cmd

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// worktreeDelay is the pause between launching Claude instances to avoid
// git worktree creation races. Tests set this to 0.
var worktreeDelay = 1 * time.Second

func runWtree(sessionName string, n int, windowTitle string) error {
	if !isGitRepo() {
		return fmt.Errorf("wtree mode requires a git repository")
	}

	if err := gitFetch("."); err != nil {
		fmt.Fprintf(os.Stderr, "warning: git fetch origin failed: %v\n", err)
	}

	reattach, err := handleExistingSession(sessionName)
	if err != nil || reattach {
		return err
	}

	runner.Run("new-session", "-d", "-s", sessionName)
	runner.Run("rename-window", "-t", sessionName+":0", windowTitle)
	runner.Run("set-window-option", "-t", sessionName+":0", "automatic-rename", "off")

	// Pane 0 is column 0 top
	col0Top, _ := runner.Run("display-message", "-t", sessionName+":0.0", "-p", "#{pane_id}")

	// Create columns 1..n-1 by splitting horizontally from column 0
	topPanes := []string{col0Top}
	for i := 1; i < n; i++ {
		pane, _ := runner.Run("split-window", "-h", "-t", col0Top, "-P", "-F", "#{pane_id}")
		topPanes = append(topPanes, pane)
	}

	// Even out column widths
	windowWidth, _ := runner.Run("display-message", "-p", "#{window_width}")
	w, _ := strconv.Atoi(windowWidth)
	if w > 0 && n > 0 {
		paneWidth := w / n
		for _, pane := range topPanes {
			runner.Run("resize-pane", "-t", pane, "-x", strconv.Itoa(paneWidth))
		}
	}

	// Split each column vertically and launch claude + lazygit pairs
	for i, topPane := range topPanes {
		if i > 0 {
			time.Sleep(worktreeDelay)
		}

		worktreeName := fmt.Sprintf("%s-%d", sessionName, i)
		worktreePath := fmt.Sprintf(".claude/worktrees/%s", worktreeName)

		// Launch claude in the top pane
		runner.Run("send-keys", "-t", topPane,
			fmt.Sprintf("clear && claude --worktree %s", worktreeName), "Enter")

		// Split vertically to create the bottom pane for lazygit
		bottomPane, _ := runner.Run("split-window", "-v", "-p", "40", "-t", topPane, "-P", "-F", "#{pane_id}")

		// Wait for worktree to be a valid git repo before launching lazygit
		runner.Run("send-keys", "-t", bottomPane,
			fmt.Sprintf("while [ ! -e %s/.git ]; do sleep 0.3; done; lazygit -p %s", worktreePath, worktreePath), "Enter")
	}

	runner.Run("select-pane", "-t", col0Top)

	return runner.Attach(sessionName)
}
