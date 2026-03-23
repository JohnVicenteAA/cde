package cmd

import "testing"

func TestAttach(t *testing.T) {
	mock := newMockTmux()
	mock.sessions["proj_ide"] = true
	orig := runner
	runner = mock
	defer func() { runner = orig }()

	attachName = "proj_ide"
	defer func() { attachName = "" }()

	err := runAttach(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.attached != "proj_ide" {
		t.Errorf("attached = %q, want %q", mock.attached, "proj_ide")
	}
}

func TestAttachNotFound(t *testing.T) {
	mock := newMockTmux()
	orig := runner
	runner = mock
	defer func() { runner = orig }()

	attachName = "nonexistent"
	defer func() { attachName = "" }()

	err := runAttach(nil, nil)
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
}
