package cmd

import (
	"fmt"
	"testing"
)

func init() {
	worktreeDelay = 0
	gitFetch = func(dir string) error { return nil }
}

func TestRunWtree(t *testing.T) {
	mock := newMockTmux()
	mock.outputs["display-message -t test_wtree:0.0 -p #{pane_id}"] = "%0"
	mock.outputs["split-window -h -t %0 -P -F #{pane_id}"] = "%1"
	mock.outputs["display-message -p #{window_width}"] = "200"
	mock.outputs["split-window -v -p 40 -t %0 -P -F #{pane_id}"] = "%10"
	mock.outputs["split-window -v -p 40 -t %1 -P -F #{pane_id}"] = "%11"
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

	// Verify window title and automatic-rename disabled
	if !mock.hasCall("rename-window", "-t", "test_wtree:0", "test: wtree") {
		t.Error("expected window to be renamed")
	}
	if !mock.hasCall("set-window-option", "-t", "test_wtree:0", "automatic-rename", "off") {
		t.Error("expected automatic-rename to be disabled")
	}

	// Verify claude launched with named worktrees in top panes
	if !mock.hasCall("send-keys", "-t", "%0", "clear && claude --worktree test_wtree-0", "Enter") {
		t.Error("expected claude --worktree test_wtree-0 in column 0 top pane")
	}
	if !mock.hasCall("send-keys", "-t", "%1", "clear && claude --worktree test_wtree-1", "Enter") {
		t.Error("expected claude --worktree test_wtree-1 in column 1 top pane")
	}

	// Verify lazygit launched in bottom panes with wait loop for worktree readiness
	if !mock.hasCall("send-keys", "-t", "%10", "while [ ! -e .claude/worktrees/test_wtree-0/.git ]; do sleep 0.3; done; lazygit -p .claude/worktrees/test_wtree-0", "Enter") {
		t.Error("expected lazygit for worktree 0 in column 0 bottom pane")
	}
	if !mock.hasCall("send-keys", "-t", "%11", "while [ ! -e .claude/worktrees/test_wtree-1/.git ]; do sleep 0.3; done; lazygit -p .claude/worktrees/test_wtree-1", "Enter") {
		t.Error("expected lazygit for worktree 1 in column 1 bottom pane")
	}

	// Verify pane resizing
	if !mock.hasCall("resize-pane", "-t", "%0", "-x", "100") {
		t.Error("expected first column resized to 100")
	}
	if !mock.hasCall("resize-pane", "-t", "%1", "-x", "100") {
		t.Error("expected second column resized to 100")
	}

	// Verify focus on first top pane
	if !mock.hasCall("select-pane", "-t", "%0") {
		t.Error("expected first top pane to be selected")
	}
}

func TestRunWtreeColumnPairing(t *testing.T) {
	for _, n := range []int{1, 2, 3} {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			mock := newMockTmux()
			mock.outputs["display-message -t sess:0.0 -p #{pane_id}"] = "%0"
			mock.outputs["display-message -p #{window_width}"] = "300"
			if n > 1 {
				seq := make([]string, 0, n-1)
				for i := 1; i < n; i++ {
					seq = append(seq, fmt.Sprintf("%%%d", i))
				}
				mock.outputSeqs["split-window -h -t %0 -P -F #{pane_id}"] = seq
			}
			for i := 0; i < n; i++ {
				mock.outputs[fmt.Sprintf("split-window -v -p 40 -t %%%d -P -F #{pane_id}", i)] = fmt.Sprintf("%%%d", 10+i)
			}
			runner = mock

			origIsGitRepo := isGitRepo
			isGitRepo = func() bool { return true }
			defer func() { isGitRepo = origIsGitRepo }()

			if err := runWtree("sess", n, "title"); err != nil {
				t.Fatalf("runWtree returned error: %v", err)
			}

			for i := 0; i < n; i++ {
				topPane := fmt.Sprintf("%%%d", i)
				bottomPane := fmt.Sprintf("%%%d", 10+i)
				worktreeName := fmt.Sprintf("sess-%d", i)
				worktreePath := fmt.Sprintf(".claude/worktrees/%s", worktreeName)

				// Each top pane gets claude with the correct worktree name
				if !mock.hasCall("send-keys", "-t", topPane, fmt.Sprintf("clear && claude --worktree %s", worktreeName), "Enter") {
					t.Errorf("column %d: expected claude --worktree %s in pane %s", i, worktreeName, topPane)
				}

				// The vertical split targets the correct top pane
				if !mock.hasCall("split-window", "-v", "-p", "40", "-t", topPane, "-P", "-F", "#{pane_id}") {
					t.Errorf("column %d: expected vertical split from top pane %s", i, topPane)
				}

				// Each bottom pane gets lazygit with wait loop for worktree readiness
				expectedLg := fmt.Sprintf("while [ ! -e %s/.git ]; do sleep 0.3; done; lazygit -p %s", worktreePath, worktreePath)
				if !mock.hasCall("send-keys", "-t", bottomPane, expectedLg, "Enter") {
					t.Errorf("column %d: expected lazygit wait+launch in pane %s", i, bottomPane)
				}
			}
		})
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

func TestRunWtreeThreeColumns(t *testing.T) {
	mock := newMockTmux()
	mock.outputs["display-message -t test_wtree:0.0 -p #{pane_id}"] = "%0"
	mock.outputSeqs["split-window -h -t %0 -P -F #{pane_id}"] = []string{"%1", "%2"}
	mock.outputs["display-message -p #{window_width}"] = "300"
	mock.outputs["split-window -v -p 40 -t %0 -P -F #{pane_id}"] = "%10"
	mock.outputs["split-window -v -p 40 -t %1 -P -F #{pane_id}"] = "%11"
	mock.outputs["split-window -v -p 40 -t %2 -P -F #{pane_id}"] = "%12"
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
	horizontalSplits := 0
	for _, c := range splits {
		if len(c.args) >= 2 && c.args[1] == "-h" {
			horizontalSplits++
		}
	}
	if horizontalSplits != 2 {
		t.Errorf("expected 2 horizontal splits, got %d", horizontalSplits)
	}

	// Verify 3 vertical splits (one per column)
	verticalSplits := 0
	for _, c := range splits {
		if len(c.args) >= 2 && c.args[1] == "-v" {
			verticalSplits++
		}
	}
	if verticalSplits != 3 {
		t.Errorf("expected 3 vertical splits, got %d", verticalSplits)
	}

	// Verify all 3 claude launches
	if !mock.hasCall("send-keys", "-t", "%0", "clear && claude --worktree test_wtree-0", "Enter") {
		t.Error("expected claude in column 0")
	}
	if !mock.hasCall("send-keys", "-t", "%1", "clear && claude --worktree test_wtree-1", "Enter") {
		t.Error("expected claude in column 1")
	}

	// Verify pane widths at 100 each
	if !mock.hasCall("resize-pane", "-t", "%0", "-x", "100") {
		t.Error("expected column 0 resized to 100")
	}
}
