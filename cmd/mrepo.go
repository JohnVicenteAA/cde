package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
)

// discoverGitRepos returns subdirectory names within dir that are git repos.
var discoverGitRepos = func(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var repos []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		gitPath := filepath.Join(dir, e.Name(), ".git")
		if _, err := os.Stat(gitPath); err == nil {
			repos = append(repos, e.Name())
		}
	}
	return repos, nil
}

// promptLabel asks the user for a short session label.
var promptLabel = func() (string, error) {
	var label string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Session label").
				Placeholder("e.g. bugfix, feature, sprint-3").
				Value(&label),
		),
	)
	if err := form.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(label), nil
}

// selectRepos prompts the user to pick repos from the list.
var selectRepos = func(repos []string) ([]string, error) {
	options := make([]huh.Option[string], len(repos))
	for i, r := range repos {
		options[i] = huh.NewOption(r, r)
	}
	var selected []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select repos to open (/ to filter)").
				Options(options...).
				Filterable(true).
				Height(15).
				Value(&selected),
		),
	)
	if err := form.Run(); err != nil {
		return nil, err
	}
	return selected, nil
}

func runMrepo() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	repos, err := discoverGitRepos(cwd)
	if err != nil {
		return fmt.Errorf("scanning for repos: %w", err)
	}
	if len(repos) == 0 {
		return fmt.Errorf("no git repositories found in %s", cwd)
	}

	selected, err := selectRepos(repos)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		return fmt.Errorf("no repos selected")
	}

	n := len(selected)

	label, err := promptLabel()
	if err != nil {
		return err
	}
	if label == "" {
		return fmt.Errorf("session label is required")
	}

	// Sanitize label for use in tmux session name
	label = strings.ReplaceAll(strings.ReplaceAll(label, ".", "_"), " ", "_")

	sessionName := "mrepo_" + label + "_" + strings.Join(selected, "_")

	reattach, err := handleExistingSession(sessionName)
	if err != nil || reattach {
		return err
	}

	title := label + " mrepo [" + strings.Join(selected, ", ") + "]"

	runner.Run("new-session", "-d", "-s", sessionName)
	runner.Run("rename-window", "-t", sessionName+":0", title)
	runner.Run("set-window-option", "-t", sessionName+":0", "automatic-rename", "off")

	col0Top, _ := runner.Run("display-message", "-t", sessionName+":0.0", "-p", "#{pane_id}")

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

	// Split each column and launch claude + lazygit per repo
	for i, topPane := range topPanes {
		if i > 0 {
			time.Sleep(worktreeDelay)
		}

		repoName := selected[i]
		repoPath := filepath.Join(cwd, repoName)
		worktreeName := fmt.Sprintf("%s-%s", label, repoName)
		worktreePath := fmt.Sprintf(".claude/worktrees/%s", worktreeName)

		// cd into repo, then launch claude with worktree
		runner.Run("send-keys", "-t", topPane,
			fmt.Sprintf("cd %s && clear && claude --worktree %s", repoPath, worktreeName), "Enter")

		// Split vertically for lazygit
		bottomPane, _ := runner.Run("split-window", "-v", "-p", "40", "-t", topPane, "-P", "-F", "#{pane_id}")

		// cd into repo, wait for worktree, launch lazygit
		runner.Run("send-keys", "-t", bottomPane,
			fmt.Sprintf("cd %s && while [ ! -e %s/.git ]; do sleep 0.3; done; lazygit -p %s",
				repoPath, worktreePath, worktreePath), "Enter")
	}

	runner.Run("select-pane", "-t", col0Top)

	return runner.Attach(sessionName)
}
