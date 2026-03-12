package cmd

import (
	"fmt"
	"strconv"
)

func runAgent(sessionName string, n int) error {
	if !isGitRepo() {
		return fmt.Errorf("agent mode requires a git repository")
	}

	reattach, err := handleExistingSession(sessionName)
	if err != nil || reattach {
		return err
	}

	runner.Run("new-session", "-d", "-s", sessionName)

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

	// Send cc -w to each top pane
	for _, pane := range topPanes {
		runner.Run("send-keys", "-t", pane, "clear && cc -w", "Enter")
	}

	runner.Run("select-pane", "-t", topID)

	return runner.Attach(sessionName)
}
