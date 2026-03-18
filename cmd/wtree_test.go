package cmd

import "testing"

func init() {
	worktreeDelay = 0
}

func TestRunWtree(t *testing.T) {
	mock := newMockTmux()
	mock.outputs["display-message -t test_wtree:0.0 -p #{pane_id}"] = "%0"
	mock.outputs["split-window -v -p 40 -t %0 -P -F #{pane_id}"] = "%1"
	mock.outputs["split-window -h -p 50 -t %1 -P -F #{pane_id}"] = "%2"
	mock.outputs["display-message -t test_wtree:0.2 -p #{pane_id}"] = "%2"
	mock.outputs["split-window -h -t %0 -P -F #{pane_id}"] = "%3"
	mock.outputs["display-message -p #{window_width}"] = "200"
	runner = mock

	origIsGitRepo := isGitRepo
	isGitRepo = func() bool { return true }
	defer func() { isGitRepo = origIsGitRepo }()

	err := runWtree("test_wtree", 2, "test: wtree")
	if err != nil {
		t.Fatalf("runWtree returned error: %v", err)
	}

	if mock.attached != "test_wtree" {
		t.Errorf("expected attach to %q, got %q", "test_wtree", mock.attached)
	}

	// Verify session created
	if !mock.hasCall("new-session", "-d", "-s", "test_wtree") {
		t.Error("expected new-session call")
	}

	// Verify bottom row: lazygit + lazydocker
	if !mock.hasCall("send-keys", "-t", "%1", "lazygit", "Enter") {
		t.Error("expected lazygit in bottom-left pane")
	}
	if !mock.hasCall("send-keys", "-t", "%2", "lazydocker", "Enter") {
		t.Error("expected lazydocker in bottom-right pane")
	}

	// Verify top panes get claude --worktree
	if !mock.hasCall("send-keys", "-t", "%0", "clear && claude --worktree", "Enter") {
		t.Error("expected claude --worktree in first top pane")
	}
	if !mock.hasCall("send-keys", "-t", "%3", "clear && claude --worktree", "Enter") {
		t.Error("expected claude --worktree in second top pane")
	}

	// Verify pane resizing
	if !mock.hasCall("resize-pane", "-t", "%0", "-x", "100") {
		t.Error("expected first pane resized to 100")
	}
	if !mock.hasCall("resize-pane", "-t", "%3", "-x", "100") {
		t.Error("expected second pane resized to 100")
	}

	// Verify window title and automatic-rename disabled
	if !mock.hasCall("rename-window", "-t", "test_wtree:0", "test: wtree") {
		t.Error("expected window to be renamed")
	}
	if !mock.hasCall("set-window-option", "-t", "test_wtree:0", "automatic-rename", "off") {
		t.Error("expected automatic-rename to be disabled")
	}

	// Verify focus on first top pane
	if !mock.hasCall("select-pane", "-t", "%0") {
		t.Error("expected first top pane to be selected")
	}
}

func TestRunWtreeRequiresGitRepo(t *testing.T) {
	mock := newMockTmux()
	runner = mock

	origIsGitRepo := isGitRepo
	isGitRepo = func() bool { return false }
	defer func() { isGitRepo = origIsGitRepo }()

	err := runWtree("test_wtree", 2, "test: wtree")
	if err == nil {
		t.Fatal("expected error when not in git repo")
	}
	if err.Error() != "wtree mode requires a git repository" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunWtreeCustomPaneCount(t *testing.T) {
	mock := newMockTmux()
	mock.outputs["display-message -t test_wtree:0.0 -p #{pane_id}"] = "%0"
	mock.outputs["split-window -v -p 40 -t %0 -P -F #{pane_id}"] = "%1"
	mock.outputs["split-window -h -p 50 -t %1 -P -F #{pane_id}"] = "%2"
	mock.outputs["display-message -t test_wtree:0.2 -p #{pane_id}"] = "%2"
	mock.outputs["split-window -h -t %0 -P -F #{pane_id}"] = "%3"
	mock.outputs["display-message -p #{window_width}"] = "300"
	runner = mock

	origIsGitRepo := isGitRepo
	isGitRepo = func() bool { return true }
	defer func() { isGitRepo = origIsGitRepo }()

	err := runWtree("test_wtree", 3, "test: wtree")
	if err != nil {
		t.Fatalf("runWtree returned error: %v", err)
	}

	// With n=3, split-window -h should be called twice
	splits := mock.findCalls("split-window")
	horizontalTopSplits := 0
	for _, c := range splits {
		if len(c.args) >= 4 && c.args[1] == "-h" && c.args[3] == "%0" {
			horizontalTopSplits++
		}
	}
	if horizontalTopSplits != 2 {
		t.Errorf("expected 2 horizontal top splits, got %d", horizontalTopSplits)
	}
}
