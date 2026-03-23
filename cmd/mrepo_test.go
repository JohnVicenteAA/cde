package cmd

import (
	"fmt"
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

func TestRunMrepo(t *testing.T) {
	defer stubNotGitRepo()()
	sn := "mrepo_test_repo-a_repo-b"
	mock := newMockTmux()
	mock.outputs[fmt.Sprintf("display-message -t %s:0.0 -p #{pane_id}", sn)] = "%0"
	mock.outputs["split-window -h -t %0 -P -F #{pane_id}"] = "%1"
	mock.outputs["display-message -p #{window_width}"] = "200"
	mock.outputs["split-window -v -p 40 -t %0 -P -F #{pane_id}"] = "%10"
	mock.outputs["split-window -v -p 40 -t %1 -P -F #{pane_id}"] = "%11"
	runner = mock

	origDiscover := discoverGitRepos
	discoverGitRepos = func(dir string) ([]string, error) {
		return []string{"repo-a", "repo-b"}, nil
	}
	defer func() { discoverGitRepos = origDiscover }()

	origSelect := selectRepos
	selectRepos = func(repos []string) ([]string, error) {
		return []string{"repo-a", "repo-b"}, nil
	}
	defer func() { selectRepos = origSelect }()

	defer stubLabel("test")()

	err := runMrepo()
	if err != nil {
		t.Fatalf("runMrepo returned error: %v", err)
	}

	if mock.attached != sn {
		t.Errorf("expected attach to %q, got %q", sn, mock.attached)
	}

	// Verify session created
	if !mock.hasCall("new-session", "-d", "-s", sn) {
		t.Error("expected new-session call")
	}

	// Verify window title includes label and selected repos
	if !mock.hasCall("rename-window", "-t", sn+":0", "test mrepo [repo-a, repo-b]") {
		t.Error("expected window to be renamed with label and repo names")
	}
	if !mock.hasCall("set-window-option", "-t", sn+":0", "automatic-rename", "off") {
		t.Error("expected automatic-rename to be disabled")
	}

	// Verify pane resizing
	if !mock.hasCall("resize-pane", "-t", "%0", "-x", "100") {
		t.Error("expected column 0 resized to 100")
	}
	if !mock.hasCall("resize-pane", "-t", "%1", "-x", "100") {
		t.Error("expected column 1 resized to 100")
	}

	// Verify focus on first top pane
	if !mock.hasCall("select-pane", "-t", "%0") {
		t.Error("expected first top pane to be selected")
	}
}

func TestRunMrepoSendKeys(t *testing.T) {
	defer stubNotGitRepo()()
	sn := "mrepo_test_repo-a_repo-b"
	mock := newMockTmux()
	mock.outputs[fmt.Sprintf("display-message -t %s:0.0 -p #{pane_id}", sn)] = "%0"
	mock.outputs["split-window -h -t %0 -P -F #{pane_id}"] = "%1"
	mock.outputs["display-message -p #{window_width}"] = "200"
	mock.outputs["split-window -v -p 40 -t %0 -P -F #{pane_id}"] = "%10"
	mock.outputs["split-window -v -p 40 -t %1 -P -F #{pane_id}"] = "%11"
	runner = mock

	origDiscover := discoverGitRepos
	discoverGitRepos = func(dir string) ([]string, error) {
		return []string{"repo-a", "repo-b"}, nil
	}
	defer func() { discoverGitRepos = origDiscover }()

	origSelect := selectRepos
	selectRepos = func(repos []string) ([]string, error) {
		return []string{"repo-a", "repo-b"}, nil
	}
	defer func() { selectRepos = origSelect }()

	defer stubLabel("test")()

	err := runMrepo()
	if err != nil {
		t.Fatalf("runMrepo returned error: %v", err)
	}

	// Verify send-keys contain cd + claude + --add-dir for top panes
	for _, c := range mock.calls {
		if len(c.args) >= 3 && c.args[0] == "send-keys" && c.args[1] == "-t" && c.args[2] == "%0" {
			if len(c.args) >= 4 {
				cmd := c.args[3]
				if contains(cmd, "claude --worktree test-repo-a") &&
					contains(cmd, "--add-dir") && contains(cmd, "repo-b/.claude/worktrees/test-repo-b") {
					goto foundClaude0
				}
			}
		}
	}
	t.Error("expected cd + claude --worktree + --add-dir for repo-b worktree in column 0")
foundClaude0:

	for _, c := range mock.calls {
		if len(c.args) >= 3 && c.args[0] == "send-keys" && c.args[1] == "-t" && c.args[2] == "%1" {
			if len(c.args) >= 4 {
				cmd := c.args[3]
				if contains(cmd, "claude --worktree test-repo-b") &&
					contains(cmd, "--add-dir") && contains(cmd, "repo-a/.claude/worktrees/test-repo-a") {
					goto foundClaude1
				}
			}
		}
	}
	t.Error("expected cd + claude --worktree + --add-dir for repo-a worktree in column 1")
foundClaude1:

	// Verify lazygit send-keys in bottom panes
	for _, c := range mock.calls {
		if len(c.args) >= 3 && c.args[0] == "send-keys" && c.args[1] == "-t" && c.args[2] == "%10" {
			if len(c.args) >= 4 {
				cmd := c.args[3]
				if contains(cmd, "lazygit -p .claude/worktrees/test-repo-a") {
					goto foundLg0
				}
			}
		}
	}
	t.Error("expected lazygit for repo-a in column 0 bottom pane")
foundLg0:

	for _, c := range mock.calls {
		if len(c.args) >= 3 && c.args[0] == "send-keys" && c.args[1] == "-t" && c.args[2] == "%11" {
			if len(c.args) >= 4 {
				cmd := c.args[3]
				if contains(cmd, "lazygit -p .claude/worktrees/test-repo-b") {
					goto foundLg1
				}
			}
		}
	}
	t.Error("expected lazygit for repo-b in column 1 bottom pane")
foundLg1:
}

func TestRunMrepoRejectsGitRepo(t *testing.T) {
	origIsGitRepo := isGitRepo
	isGitRepo = func() bool { return true }
	defer func() { isGitRepo = origIsGitRepo }()

	err := runMrepo()
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
	discoverGitRepos = func(dir string) ([]string, error) {
		return nil, nil
	}
	defer func() { discoverGitRepos = origDiscover }()

	err := runMrepo()
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
	discoverGitRepos = func(dir string) ([]string, error) {
		return []string{"repo-a"}, nil
	}
	defer func() { discoverGitRepos = origDiscover }()

	origSelect := selectRepos
	selectRepos = func(repos []string) ([]string, error) {
		return nil, nil
	}
	defer func() { selectRepos = origSelect }()

	err := runMrepo()
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
	discoverGitRepos = func(dir string) ([]string, error) {
		return []string{"repo-a"}, nil
	}
	defer func() { discoverGitRepos = origDiscover }()

	origSelect := selectRepos
	selectRepos = func(repos []string) ([]string, error) {
		return []string{"repo-a"}, nil
	}
	defer func() { selectRepos = origSelect }()

	defer stubLabel("")()

	err := runMrepo()
	if err == nil {
		t.Fatal("expected error when label is empty")
	}
	if err.Error() != "session label is required" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunMrepoSingleRepo(t *testing.T) {
	defer stubNotGitRepo()()
	sn := "mrepo_test_solo-repo"
	mock := newMockTmux()
	mock.outputs[fmt.Sprintf("display-message -t %s:0.0 -p #{pane_id}", sn)] = "%0"
	mock.outputs["display-message -p #{window_width}"] = "200"
	mock.outputs["split-window -v -p 40 -t %0 -P -F #{pane_id}"] = "%10"
	runner = mock

	origDiscover := discoverGitRepos
	discoverGitRepos = func(dir string) ([]string, error) {
		return []string{"solo-repo"}, nil
	}
	defer func() { discoverGitRepos = origDiscover }()

	origSelect := selectRepos
	selectRepos = func(repos []string) ([]string, error) {
		return []string{"solo-repo"}, nil
	}
	defer func() { selectRepos = origSelect }()

	defer stubLabel("test")()

	err := runMrepo()
	if err != nil {
		t.Fatalf("runMrepo returned error: %v", err)
	}

	// With 1 repo, no horizontal splits
	splits := mock.findCalls("split-window")
	for _, c := range splits {
		if len(c.args) >= 2 && c.args[1] == "-h" {
			t.Error("expected no horizontal splits for single repo")
		}
	}

	// Single repo should NOT have --add-dir
	for _, c := range mock.calls {
		if len(c.args) >= 4 && c.args[0] == "send-keys" && contains(c.args[3], "claude") {
			if contains(c.args[3], "--add-dir") {
				t.Error("single repo should not have --add-dir flags")
			}
		}
	}

	// Should still have 1 vertical split
	verticalSplits := 0
	for _, c := range splits {
		if len(c.args) >= 2 && c.args[1] == "-v" {
			verticalSplits++
		}
	}
	if verticalSplits != 1 {
		t.Errorf("expected 1 vertical split, got %d", verticalSplits)
	}
}

func TestRunMrepoThreeRepos(t *testing.T) {
	defer stubNotGitRepo()()
	sn := "mrepo_test_alpha_beta_gamma"
	mock := newMockTmux()
	mock.outputs[fmt.Sprintf("display-message -t %s:0.0 -p #{pane_id}", sn)] = "%0"
	mock.outputSeqs["split-window -h -t %0 -P -F #{pane_id}"] = []string{"%1", "%2"}
	mock.outputs["display-message -p #{window_width}"] = "300"
	for i := 0; i < 3; i++ {
		mock.outputs[fmt.Sprintf("split-window -v -p 40 -t %%%d -P -F #{pane_id}", i)] = fmt.Sprintf("%%%d", 10+i)
	}
	runner = mock

	repos := []string{"alpha", "beta", "gamma"}

	origDiscover := discoverGitRepos
	discoverGitRepos = func(dir string) ([]string, error) { return repos, nil }
	defer func() { discoverGitRepos = origDiscover }()

	origSelect := selectRepos
	selectRepos = func(r []string) ([]string, error) { return repos, nil }
	defer func() { selectRepos = origSelect }()

	defer stubLabel("test")()

	if err := runMrepo(); err != nil {
		t.Fatalf("runMrepo returned error: %v", err)
	}

	// 2 horizontal splits for 3 columns
	splits := mock.findCalls("split-window")
	hSplits := 0
	vSplits := 0
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
	if hSplits != 2 {
		t.Errorf("expected 2 horizontal splits, got %d", hSplits)
	}
	if vSplits != 3 {
		t.Errorf("expected 3 vertical splits, got %d", vSplits)
	}

	// Verify column widths at 100 each
	if !mock.hasCall("resize-pane", "-t", "%0", "-x", "100") {
		t.Error("expected column 0 resized to 100")
	}
}

func TestRunMrepoAddDirThreeRepos(t *testing.T) {
	defer stubNotGitRepo()()
	sn := "mrepo_test_alpha_beta_gamma"
	mock := newMockTmux()
	mock.outputs[fmt.Sprintf("display-message -t %s:0.0 -p #{pane_id}", sn)] = "%0"
	mock.outputSeqs["split-window -h -t %0 -P -F #{pane_id}"] = []string{"%1", "%2"}
	mock.outputs["display-message -p #{window_width}"] = "300"
	for i := 0; i < 3; i++ {
		mock.outputs[fmt.Sprintf("split-window -v -p 40 -t %%%d -P -F #{pane_id}", i)] = fmt.Sprintf("%%%d", 10+i)
	}
	runner = mock

	repos := []string{"alpha", "beta", "gamma"}

	origDiscover := discoverGitRepos
	discoverGitRepos = func(dir string) ([]string, error) { return repos, nil }
	defer func() { discoverGitRepos = origDiscover }()

	origSelect := selectRepos
	selectRepos = func(r []string) ([]string, error) { return repos, nil }
	defer func() { selectRepos = origSelect }()

	defer stubLabel("test")()

	if err := runMrepo(); err != nil {
		t.Fatalf("runMrepo returned error: %v", err)
	}

	// Collect claude send-keys per pane
	claudeCmds := map[string]string{}
	for _, c := range mock.calls {
		if len(c.args) >= 4 && c.args[0] == "send-keys" && contains(c.args[3], "claude") {
			claudeCmds[c.args[2]] = c.args[3]
		}
	}

	// alpha (%0) should have --add-dir for beta and gamma worktrees
	if cmd, ok := claudeCmds["%0"]; !ok {
		t.Error("no claude send-keys for %0")
	} else {
		if !contains(cmd, "beta/.claude/worktrees/test-beta") {
			t.Error("alpha column missing --add-dir for beta worktree")
		}
		if !contains(cmd, "gamma/.claude/worktrees/test-gamma") {
			t.Error("alpha column missing --add-dir for gamma worktree")
		}
		if contains(cmd, "alpha/.claude/worktrees/test-alpha") {
			t.Error("alpha column should not --add-dir itself")
		}
	}

	// beta (%1) should have --add-dir for alpha and gamma worktrees
	if cmd, ok := claudeCmds["%1"]; !ok {
		t.Error("no claude send-keys for %1")
	} else {
		if !contains(cmd, "alpha/.claude/worktrees/test-alpha") {
			t.Error("beta column missing --add-dir for alpha worktree")
		}
		if !contains(cmd, "gamma/.claude/worktrees/test-gamma") {
			t.Error("beta column missing --add-dir for gamma worktree")
		}
	}

	// gamma (%2) should have --add-dir for alpha and beta worktrees
	if cmd, ok := claudeCmds["%2"]; !ok {
		t.Error("no claude send-keys for %2")
	} else {
		if !contains(cmd, "alpha/.claude/worktrees/test-alpha") {
			t.Error("gamma column missing --add-dir for alpha worktree")
		}
		if !contains(cmd, "beta/.claude/worktrees/test-beta") {
			t.Error("gamma column missing --add-dir for beta worktree")
		}
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
