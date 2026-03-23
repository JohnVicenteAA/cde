package cmd

import (
	"testing"
)

func TestIsCdeSession(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"myproject_ide", true},
		{"myproject_wtree", true},
		{"mrepo_label_repo1", true},
		{"mrepo_x", true},
		{"random-session", false},
		{"ide_stuff", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCdeSession(tt.name); got != tt.want {
				t.Errorf("isCdeSession(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestListCdeSessions(t *testing.T) {
	mock := newMockTmux()
	mock.outputs["list-sessions -F #{session_name}"] = "myproject_ide\nrandom\nmrepo_label_repo1\nfoo_wtree"
	orig := runner
	runner = mock
	defer func() { runner = orig }()

	sessions, err := listCdeSessions()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"myproject_ide", "mrepo_label_repo1", "foo_wtree"}
	if len(sessions) != len(want) {
		t.Fatalf("got %d sessions, want %d: %v", len(sessions), len(want), sessions)
	}
	for i := range want {
		if sessions[i] != want[i] {
			t.Errorf("session[%d] = %q, want %q", i, sessions[i], want[i])
		}
	}
}

func TestRunSessionListNoSessions(t *testing.T) {
	mock := newMockTmux()
	mock.outputs["list-sessions -F #{session_name}"] = ""
	orig := runner
	runner = mock
	defer func() { runner = orig }()

	listSessions = true
	defer func() { listSessions = false }()

	err := runSession(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunSessionList(t *testing.T) {
	mock := newMockTmux()
	mock.outputs["list-sessions -F #{session_name}"] = "proj_ide\nproj_wtree\nrandom"
	orig := runner
	runner = mock
	defer func() { runner = orig }()

	listSessions = true
	defer func() { listSessions = false }()

	err := runSession(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteByName(t *testing.T) {
	mock := newMockTmux()
	mock.sessions["proj_ide"] = true
	orig := runner
	runner = mock
	defer func() { runner = orig }()

	deleteName = "proj_ide"
	deleteAll = false
	defer func() { deleteName = ""; deleteAll = false }()

	err := runSessionDelete(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.hasCall("kill-session", "-t", "=proj_ide") {
		t.Error("expected kill-session call for proj_ide")
	}
}

func TestDeleteByNameNotFound(t *testing.T) {
	mock := newMockTmux()
	orig := runner
	runner = mock
	defer func() { runner = orig }()

	deleteName = "nonexistent"
	deleteAll = false
	defer func() { deleteName = ""; deleteAll = false }()

	err := runSessionDelete(nil, nil)
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
}

func TestDeleteAll(t *testing.T) {
	mock := newMockTmux()
	mock.outputs["list-sessions -F #{session_name}"] = "proj_ide\nproj_wtree\nrandom"
	orig := runner
	runner = mock
	defer func() { runner = orig }()

	deleteName = ""
	deleteAll = true
	defer func() { deleteName = ""; deleteAll = false }()

	err := runSessionDelete(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.hasCall("kill-session", "-t", "=proj_ide") {
		t.Error("expected kill-session call for proj_ide")
	}
	if !mock.hasCall("kill-session", "-t", "=proj_wtree") {
		t.Error("expected kill-session call for proj_wtree")
	}
	if mock.hasCall("kill-session", "-t", "=random") {
		t.Error("should not kill non-cde session")
	}
}

func TestDeleteNoFlags(t *testing.T) {
	deleteName = ""
	deleteAll = false

	err := runSessionDelete(nil, nil)
	if err == nil {
		t.Fatal("expected error when no flags provided")
	}
}
