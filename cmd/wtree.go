package cmd

import (
	"fmt"
	"strconv"
	"time"
)

// worktreeDelay is the pause between launching Claude instances to avoid
// git worktree creation races. Tests set this to 0.
var worktreeDelay = 3 * time.Second

func runWtree(sessionName string, n int, windowTitle string) error {
	if !isGitRepo() {
		return fmt.Errorf("wtree mode requires a git repository")
	}

	reattach, err := handleExistingSession(sessionName)
	if err != nil || reattach {
		return err
	}

	runner.Run("new-session", "-d", "-s", sessionName)
	runner.Run("rename-window", "-t", sessionName+":0", windowTitle)
	runner.Run("set-window-option", "-t", sessionName+":0", "automatic-rename", "off")

	// Get the initial pane ID
	topID, _ := runner.Run("display-message", "-t", sessionName+":0.0", "-p", "#{pane_id}")

	// Create bottom row (40%)
	bottomID, _ := runner.Run("split-window", "-v", "-p", "40", "-t", topID, "-P", "-F", "#{pane_id}")

	// Split bottom into lazygit | lazydocker
	runner.Run("split-window", "-h", "-p", "50", "-t", bottomID, "-P", "-F", "#{pane_id}")
	runner.Run("send-keys", "-t", bottomID, "lazygit", "Enter")
	// Bottom right is the next pane after bottomID — target via index offset
	bottomRightID, _ := runner.Run("display-message", "-t", sessionName+":0.2", "-p", "#{pane_id}")
	runner.Run("send-keys", "-t", bottomRightID, "lazydocker", "Enter")

	// Split top row into N panes
	topPanes := []string{topID}
	for i := 1; i < n; i++ {
		newPane, _ := runner.Run("split-window", "-h", "-t", topID, "-P", "-F", "#{pane_id}")
		topPanes = append(topPanes, newPane)
	}

	// Even out top pane widths
	windowWidth, _ := runner.Run("display-message", "-p", "#{window_width}")
	w, _ := strconv.Atoi(windowWidth)
	if w > 0 && n > 0 {
		paneWidth := w / n
		for _, pane := range topPanes {
			runner.Run("resize-pane", "-t", pane, "-x", strconv.Itoa(paneWidth))
		}
	}

	// Send cc -w to each top pane, staggering launches to avoid
	// git worktree creation races that cause Claude to hang
	for i, pane := range topPanes {
		if i > 0 {
			time.Sleep(worktreeDelay)
		}
		runner.Run("send-keys", "-t", pane, "clear && claude --worktree", "Enter")
	}

	runner.Run("select-pane", "-t", topID)

	return runner.Attach(sessionName)
}
