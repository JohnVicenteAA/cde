package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

func runMrepo(flagLabel string, flagRepos []string) error {
	if isGitRepo() {
		return fmt.Errorf("mrepo mode should be run from a parent directory containing repos, not from inside a git repo (use wtree mode instead)")
	}

	hasLabel := flagLabel != ""
	hasRepos := len(flagRepos) > 0
	if hasLabel != hasRepos {
		return fmt.Errorf("--label and --repo must both be provided for non-interactive mrepo")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	var selected []string
	if hasRepos {
		for _, r := range flagRepos {
			gitPath := filepath.Join(cwd, r, ".git")
			if _, err := os.Stat(gitPath); err != nil {
				return fmt.Errorf("repo %q is not a git repository in %s", r, cwd)
			}
		}
		selected = flagRepos
	} else {
		repos, err := discoverGitRepos(cwd)
		if err != nil {
			return fmt.Errorf("scanning for repos: %w", err)
		}
		if len(repos) == 0 {
			return fmt.Errorf("no git repositories found in %s", cwd)
		}

		selected, err = selectRepos(repos)
		if err != nil {
			return err
		}
		if len(selected) == 0 {
			return fmt.Errorf("no repos selected")
		}
	}

	var label string
	if hasLabel {
		label = strings.TrimSpace(flagLabel)
	} else {
		label, err = promptLabel()
		if err != nil {
			return err
		}
	}
	if label == "" {
		return fmt.Errorf("session label is required")
	}

	// Sanitize label for use in tmux session name and branch names
	label = strings.ReplaceAll(strings.ReplaceAll(label, ".", "_"), " ", "_")

	// Fetch latest from origin for each selected repo
	for _, repoName := range selected {
		repoPath := filepath.Join(cwd, repoName)
		if err := gitFetch(repoPath); err != nil {
			fmt.Fprintf(os.Stderr, "warning: git fetch origin failed for %s: %v\n", repoName, err)
		}
	}

	sessionName := "mrepo_" + label + "_" + strings.Join(selected, "_")

	reattach, err := handleExistingSession(sessionName)
	if err != nil || reattach {
		return err
	}

	title := label + " mrepo [" + strings.Join(selected, ", ") + "]"
	n := len(selected)

	runner.Run("new-session", "-d", "-s", sessionName)
	runner.Run("rename-window", "-t", sessionName+":0", title)
	runner.Run("set-window-option", "-t", sessionName+":0", "automatic-rename", "off")

	// Top pane is the single claude session
	topPane, _ := runner.Run("display-message", "-t", sessionName+":0.0", "-p", "#{pane_id}")

	// Create bottom row (40% height)
	bottomPane, _ := runner.Run("split-window", "-v", "-p", "40", "-t", topPane, "-P", "-F", "#{pane_id}")

	// Split bottom row into N panes (one per repo)
	bottomPanes := []string{bottomPane}
	for i := 1; i < n; i++ {
		pane, _ := runner.Run("split-window", "-h", "-t", bottomPane, "-P", "-F", "#{pane_id}")
		bottomPanes = append(bottomPanes, pane)
	}

	// Even out bottom pane widths
	windowWidth, _ := runner.Run("display-message", "-p", "#{window_width}")
	w, _ := strconv.Atoi(windowWidth)
	if w > 0 && n > 0 {
		paneWidth := w / n
		for _, pane := range bottomPanes {
			runner.Run("resize-pane", "-t", pane, "-x", strconv.Itoa(paneWidth))
		}
	}

	// Launch claude in top pane with --add-dir for each repo
	var addDirs string
	for _, repoName := range selected {
		repoPath := filepath.Join(cwd, repoName)
		addDirs += fmt.Sprintf(" --add-dir %s", repoPath)
	}
	runner.Run("send-keys", "-t", topPane,
		fmt.Sprintf("cd %s && clear && claude%s", cwd, addDirs), "Enter")

	// Launch lazygit in each bottom pane via a worktree on a new branch
	for i, pane := range bottomPanes {
		repoName := selected[i]
		repoPath := filepath.Join(cwd, repoName)
		branchName := fmt.Sprintf("%s/%s", label, repoName)
		worktreePath := filepath.Join(repoPath, ".claude", "worktrees", fmt.Sprintf("%s-%s", label, repoName))
		runner.Run("send-keys", "-t", pane,
			fmt.Sprintf("cd %s && git worktree add -b %s %s origin/main && lazygit -p %s",
				repoPath, branchName, worktreePath, worktreePath), "Enter")
	}

	runner.Run("select-pane", "-t", topPane)

	return runner.Attach(sessionName)
}
