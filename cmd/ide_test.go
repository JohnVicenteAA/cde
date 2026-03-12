package cmd

import "testing"

func TestRunIDE(t *testing.T) {
	mock := newMockTmux()
	runner = mock

	err := runIDE("test_ide", "test: ide")
	if err != nil {
		t.Fatalf("runIDE returned error: %v", err)
	}

	if mock.attached != "test_ide" {
		t.Errorf("expected attach to %q, got %q", "test_ide", mock.attached)
	}

	if !mock.hasCall("new-session", "-d", "-s", "test_ide") {
		t.Error("expected new-session call")
	}

	if !mock.hasCall("split-window", "-h", "-t", "test_ide:0") {
		t.Error("expected horizontal split")
	}

	if !mock.hasCall("split-window", "-v", "-t", "test_ide:0.1") {
		t.Error("expected vertical split")
	}

	if !mock.hasCall("send-keys", "-t", "test_ide:0.0", "nvim", "Enter") {
		t.Error("expected nvim to be launched in pane 0")
	}

	if !mock.hasCall("select-pane", "-t", "test_ide:0.0") {
		t.Error("expected pane 0 to be selected")
	}

	if !mock.hasCall("rename-window", "-t", "test_ide:0", "test: ide") {
		t.Error("expected window to be renamed")
	}

	if !mock.hasCall("set-window-option", "-t", "test_ide:0", "automatic-rename", "off") {
		t.Error("expected automatic-rename to be disabled")
	}
}
