package cmd

import (
	"fmt"
	"os"
	"testing"
)

func stubNotGitRepo() func() {
	orig := isGitRepo
	isGitRepo = func() bool { return false }
	return func() { isGitRepo = orig }
}

func stubLabel(label string) func() {
	orig := promptLabel
	promptLabel = func() (string, error) { return label, nil }
	return func() { promptLabel = orig }
}

func stubGitFetch() func() {
	orig := gitFetch
	gitFetch = func(dir string) error { return nil }
	return func() { gitFetch = orig }
}

func setupMrepoMock(sn string, repos []string) *mockTmux {
	mock := newMockTmux()
	mock.outputs[fmt.Sprintf("display-message -t %s:0.0 -p #{pane_id}", sn)] = "%0"
	mock.outputs["split-window -v -p 40 -t %0 -P -F #{pane_id}"] = "%10"
	mock.outputs["display-message -p #{window_width}"] = "200"

	n := len(repos)
	if n > 1 {
		seqs := make([]string, n-1)
		for i := range seqs {
			seqs[i] = fmt.Sprintf("%%%d", 11+i)
		}
		mock.outputSeqs["split-window -h -t %10 -P -F #{pane_id}"] = seqs
	}

	runner = mock
	return mock
}

func TestRunMrepo(t *testing.T) {
	defer stubNotGitRepo()()
	defer stubGitFetch()()
	defer stubLabel("test")()

	repos := []string{"repo-a", "repo-b"}
	sn := "mrepo_test_repo-a_repo-b"
	mock := setupMrepoMock(sn, repos)

	origDiscover := discoverGitRepos
	discoverGitRepos = func(dir string) ([]string, error) { return repos, nil }
	defer func() { discoverGitRepos = origDiscover }()

	origSelect := selectRepos
	selectRepos = func(r []string) ([]string, error) { return repos, nil }
	defer func() { selectRepos = origSelect }()

	if err := runMrepo("", nil); err != nil {
		t.Fatalf("runMrepo returned error: %v", err)
	}

	if mock.attached != sn {
		t.Errorf("expected attach to %q, got %q", sn, mock.attached)
	}

	if !mock.hasCall("new-session", "-d", "-s", sn) {
		t.Error("expected new-session call")
	}
	if !mock.hasCall("rename-window", "-t", sn+":0", "test mrepo [repo-a, repo-b]") {
		t.Error("expected window to be renamed with label and repo names")
	}
	if !mock.hasCall("set-window-option", "-t", sn+":0", "automatic-rename", "off") {
		t.Error("expected automatic-rename to be disabled")
	}

	// One vertical split (top/bottom), one horizontal split (2 bottom panes)
	splits := mock.findCalls("split-window")
	vSplits, hSplits := 0, 0
	for _, c := range splits {
		if len(c.args) >= 2 {
			if c.args[1] == "-v" {
				vSplits++
			}
			if c.args[1] == "-h" {
				hSplits++
			}
		}
	}
	if vSplits != 1 {
		t.Errorf("expected 1 vertical split, got %d", vSplits)
	}
	if hSplits != 1 {
		t.Errorf("expected 1 horizontal split, got %d", hSplits)
	}

	// Bottom panes resized
	if !mock.hasCall("resize-pane", "-t", "%10", "-x", "100") {
		t.Error("expected bottom pane 0 resized to 100")
	}
	if !mock.hasCall("resize-pane", "-t", "%11", "-x", "100") {
		t.Error("expected bottom pane 1 resized to 100")
	}

	if !mock.hasCall("select-pane", "-t", "%0") {
		t.Error("expected top pane to be selected")
	}
}

func TestRunMrepoSendKeys(t *testing.T) {
	defer stubNotGitRepo()()
	defer stubGitFetch()()
	defer stubLabel("test")()

	repos := []string{"repo-a", "repo-b"}
	sn := "mrepo_test_repo-a_repo-b"
	mock := setupMrepoMock(sn, repos)

	origDiscover := discoverGitRepos
	discoverGitRepos = func(dir string) ([]string, error) { return repos, nil }
	defer func() { discoverGitRepos = origDiscover }()

	origSelect := selectRepos
	selectRepos = func(r []string) ([]string, error) { return repos, nil }
	defer func() { selectRepos = origSelect }()

	if err := runMrepo("", nil); err != nil {
		t.Fatalf("runMrepo returned error: %v", err)
	}

	// Verify claude send-keys in top pane has --add-dir for both repos
	foundClaude := false
	for _, c := range mock.calls {
		if len(c.args) >= 4 && c.args[0] == "send-keys" && c.args[2] == "%0" {
			cmd := c.args[3]
			if contains(cmd, "claude") {
				foundClaude = true
				if !contains(cmd, "--add-dir") {
					t.Error("claude command missing --add-dir flags")
				}
				if !contains(cmd, "repo-a") {
					t.Error("claude command missing --add-dir for repo-a")
				}
				if !contains(cmd, "repo-b") {
					t.Error("claude command missing --add-dir for repo-b")
				}
				if contains(cmd, "--worktree") {
					t.Error("claude command should not use --worktree")
				}
			}
		}
	}
	if !foundClaude {
		t.Error("no claude send-keys found for top pane %0")
	}

	// Verify lazygit send-keys in bottom panes with git checkout -b
	for _, paneID := range []string{"%10", "%11"} {
		found := false
		for _, c := range mock.calls {
			if len(c.args) >= 4 && c.args[0] == "send-keys" && c.args[2] == paneID {
				cmd := c.args[3]
				if contains(cmd, "lazygit") && contains(cmd, "git worktree add -b test/") {
					found = true
				}
			}
		}
		if !found {
			t.Errorf("expected lazygit with git checkout -b in pane %s", paneID)
		}
	}
}

func TestRunMrepoRejectsGitRepo(t *testing.T) {
	origIsGitRepo := isGitRepo
	isGitRepo = func() bool { return true }
	defer func() { isGitRepo = origIsGitRepo }()

	err := runMrepo("", nil)
	if err == nil {
		t.Fatal("expected error when run from inside a git repo")
	}
	if !contains(err.Error(), "not from inside a git repo") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunMrepoNoReposFound(t *testing.T) {
	defer stubNotGitRepo()()
	origDiscover := discoverGitRepos
	discoverGitRepos = func(dir string) ([]string, error) { return nil, nil }
	defer func() { discoverGitRepos = origDiscover }()

	err := runMrepo("", nil)
	if err == nil {
		t.Fatal("expected error when no repos found")
	}
	if !contains(err.Error(), "no git repositories found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunMrepoNoneSelected(t *testing.T) {
	defer stubNotGitRepo()()
	origDiscover := discoverGitRepos
	discoverGitRepos = func(dir string) ([]string, error) { return []string{"repo-a"}, nil }
	defer func() { discoverGitRepos = origDiscover }()

	origSelect := selectRepos
	selectRepos = func(repos []string) ([]string, error) { return nil, nil }
	defer func() { selectRepos = origSelect }()

	err := runMrepo("", nil)
	if err == nil {
		t.Fatal("expected error when no repos selected")
	}
	if err.Error() != "no repos selected" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunMrepoEmptyLabel(t *testing.T) {
	defer stubNotGitRepo()()
	origDiscover := discoverGitRepos
	discoverGitRepos = func(dir string) ([]string, error) { return []string{"repo-a"}, nil }
	defer func() { discoverGitRepos = origDiscover }()

	origSelect := selectRepos
	selectRepos = func(repos []string) ([]string, error) { return []string{"repo-a"}, nil }
	defer func() { selectRepos = origSelect }()

	defer stubLabel("")()

	err := runMrepo("", nil)
	if err == nil {
		t.Fatal("expected error when label is empty")
	}
	if err.Error() != "session label is required" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunMrepoSingleRepo(t *testing.T) {
	defer stubNotGitRepo()()
	defer stubGitFetch()()
	defer stubLabel("test")()

	repos := []string{"solo-repo"}
	sn := "mrepo_test_solo-repo"
	mock := setupMrepoMock(sn, repos)

	origDiscover := discoverGitRepos
	discoverGitRepos = func(dir string) ([]string, error) { return repos, nil }
	defer func() { discoverGitRepos = origDiscover }()

	origSelect := selectRepos
	selectRepos = func(r []string) ([]string, error) { return repos, nil }
	defer func() { selectRepos = origSelect }()

	if err := runMrepo("", nil); err != nil {
		t.Fatalf("runMrepo returned error: %v", err)
	}

	// No horizontal splits — single bottom pane
	splits := mock.findCalls("split-window")
	for _, c := range splits {
		if len(c.args) >= 2 && c.args[1] == "-h" {
			t.Error("expected no horizontal splits for single repo")
		}
	}

	// Claude should still have --add-dir for the single repo
	for _, c := range mock.calls {
		if len(c.args) >= 4 && c.args[0] == "send-keys" && c.args[2] == "%0" {
			if contains(c.args[3], "claude") && !contains(c.args[3], "--add-dir") {
				t.Error("claude should have --add-dir even for single repo")
			}
		}
	}
}

func TestRunMrepoThreeRepos(t *testing.T) {
	defer stubNotGitRepo()()
	defer stubGitFetch()()
	defer stubLabel("test")()

	repos := []string{"alpha", "beta", "gamma"}
	sn := "mrepo_test_alpha_beta_gamma"
	mock := setupMrepoMock(sn, repos)

	origDiscover := discoverGitRepos
	discoverGitRepos = func(dir string) ([]string, error) { return repos, nil }
	defer func() { discoverGitRepos = origDiscover }()

	origSelect := selectRepos
	selectRepos = func(r []string) ([]string, error) { return repos, nil }
	defer func() { selectRepos = origSelect }()

	if err := runMrepo("", nil); err != nil {
		t.Fatalf("runMrepo returned error: %v", err)
	}

	splits := mock.findCalls("split-window")
	hSplits, vSplits := 0, 0
	for _, c := range splits {
		if len(c.args) >= 2 {
			if c.args[1] == "-h" {
				hSplits++
			}
			if c.args[1] == "-v" {
				vSplits++
			}
		}
	}
	if vSplits != 1 {
		t.Errorf("expected 1 vertical split, got %d", vSplits)
	}
	if hSplits != 2 {
		t.Errorf("expected 2 horizontal splits, got %d", hSplits)
	}

	// Verify branch names in lazygit commands
	for _, c := range mock.calls {
		if len(c.args) >= 4 && c.args[0] == "send-keys" {
			cmd := c.args[3]
			if contains(cmd, "lazygit") {
				if !contains(cmd, "git worktree add -b test/") {
					t.Errorf("expected git worktree add -b test/<repo> in: %s", cmd)
				}
			}
		}
	}
}

func TestRunMrepoNonInteractive(t *testing.T) {
	defer stubNotGitRepo()()
	defer stubGitFetch()()

	repos := []string{"repo-a", "repo-b"}
	sn := "mrepo_bugfix_repo-a_repo-b"
	mock := setupMrepoMock(sn, repos)

	// Create temp dirs that look like git repos
	dir := t.TempDir()
	for _, r := range repos {
		repoPath := fmt.Sprintf("%s/%s/.git", dir, r)
		if err := os.MkdirAll(repoPath, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	err := runMrepo("bugfix", []string{"repo-a", "repo-b"})
	if err != nil {
		t.Fatalf("runMrepo returned error: %v", err)
	}

	if mock.attached != sn {
		t.Errorf("expected attach to %q, got %q", sn, mock.attached)
	}
	if !mock.hasCall("new-session", "-d", "-s", sn) {
		t.Error("expected new-session call")
	}
	if !mock.hasCall("rename-window", "-t", sn+":0", "bugfix mrepo [repo-a, repo-b]") {
		t.Error("expected window renamed with label and repos")
	}
}

func TestRunMrepoLabelWithoutRepos(t *testing.T) {
	defer stubNotGitRepo()()
	err := runMrepo("bugfix", nil)
	if err == nil {
		t.Fatal("expected error when --label provided without --repo")
	}
	if !contains(err.Error(), "--label and --repo must both be provided") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunMrepoReposWithoutLabel(t *testing.T) {
	defer stubNotGitRepo()()
	err := runMrepo("", []string{"repo-a"})
	if err == nil {
		t.Fatal("expected error when --repo provided without --label")
	}
	if !contains(err.Error(), "--label and --repo must both be provided") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunMrepoNonInteractiveInvalidRepo(t *testing.T) {
	defer stubNotGitRepo()()
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	err := runMrepo("bugfix", []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent repo")
	}
	if !contains(err.Error(), "not a git repository") {
		t.Errorf("unexpected error: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
