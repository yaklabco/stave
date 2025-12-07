package changelog

import (
	"testing"
)

// mockGitOps implements GitOps for testing.
type mockGitOps struct {
	changedFiles  []string
	changedErr    error
	mergeBase     string
	mergeBaseErr  error
	refExists     map[string]bool
	currentBranch string
	branchErr     error
}

func (m *mockGitOps) ChangedFiles(_, _ string) ([]string, error) {
	if m.changedErr != nil {
		return nil, m.changedErr
	}
	return m.changedFiles, nil
}

func (m *mockGitOps) MergeBase(_, _ string) (string, error) {
	if m.mergeBaseErr != nil {
		return "", m.mergeBaseErr
	}
	return m.mergeBase, nil
}

func (m *mockGitOps) RefExists(ref string) bool {
	if m.refExists == nil {
		return false
	}
	return m.refExists[ref]
}

func (m *mockGitOps) CurrentBranch() (string, error) {
	if m.branchErr != nil {
		return "", m.branchErr
	}
	return m.currentBranch, nil
}

func TestContainsFile(t *testing.T) {
	tests := []struct {
		name   string
		files  []string
		target string
		want   bool
	}{
		{
			name:   "file exists",
			files:  []string{"README.md", "CHANGELOG.md", "main.go"},
			target: "CHANGELOG.md",
			want:   true,
		},
		{
			name:   "file not found",
			files:  []string{"README.md", "main.go"},
			target: "CHANGELOG.md",
			want:   false,
		},
		{
			name:   "empty list",
			files:  []string{},
			target: "CHANGELOG.md",
			want:   false,
		},
		{
			name:   "nil list",
			files:  nil,
			target: "CHANGELOG.md",
			want:   false,
		},
		{
			name:   "case sensitive",
			files:  []string{"changelog.md"},
			target: "CHANGELOG.md",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsFile(tt.files, tt.target); got != tt.want {
				t.Errorf("ContainsFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMockGitOps verifies the mock behaves correctly.
func TestMockGitOps(t *testing.T) {
	mock := &mockGitOps{
		changedFiles:  []string{"file1.go", "file2.go"},
		mergeBase:     "abc123",
		refExists:     map[string]bool{"refs/heads/main": true},
		currentBranch: "feature-branch",
	}

	t.Run("ChangedFiles", func(t *testing.T) {
		files, err := mock.ChangedFiles("base", "head")
		if err != nil {
			t.Fatalf("ChangedFiles() error = %v", err)
		}
		if len(files) != 2 {
			t.Errorf("ChangedFiles() returned %d files, want 2", len(files))
		}
	})

	t.Run("MergeBase", func(t *testing.T) {
		base, err := mock.MergeBase("ref1", "ref2")
		if err != nil {
			t.Fatalf("MergeBase() error = %v", err)
		}
		if base != "abc123" {
			t.Errorf("MergeBase() = %q, want abc123", base)
		}
	})

	t.Run("RefExists", func(t *testing.T) {
		if !mock.RefExists("refs/heads/main") {
			t.Error("RefExists(refs/heads/main) = false, want true")
		}
		if mock.RefExists("refs/heads/other") {
			t.Error("RefExists(refs/heads/other) = true, want false")
		}
	})

	t.Run("CurrentBranch", func(t *testing.T) {
		branch, err := mock.CurrentBranch()
		if err != nil {
			t.Fatalf("CurrentBranch() error = %v", err)
		}
		if branch != "feature-branch" {
			t.Errorf("CurrentBranch() = %q, want feature-branch", branch)
		}
	})
}
