package cmd

import "testing"

func TestSessionName(t *testing.T) {
	tests := []struct {
		name string
		mode string
		want string
	}{
		{"myproject", "ide", "myproject_ide"},
		{"myproject", "wtree", "myproject_wtree"},
		{"my.project", "ide", "my_project_ide"},
		{"a.b.c", "wtree", "a_b_c_wtree"},
		{"nodots", "ide", "nodots_ide"},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_"+tt.mode, func(t *testing.T) {
			got := sessionName(tt.name, tt.mode)
			if got != tt.want {
				t.Errorf("sessionName(%q, %q) = %q, want %q", tt.name, tt.mode, got, tt.want)
			}
		})
	}
}
