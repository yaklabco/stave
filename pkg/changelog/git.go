package changelog

import (
	"strings"

	"github.com/yaklabco/stave/pkg/sh"
)

// GitOps abstracts git operations for testability.
type GitOps interface {
	// ChangedFiles returns files changed between base and head.
	ChangedFiles(base, head string) ([]string, error)
	// MergeBase finds the merge base between two refs.
	MergeBase(ref1, ref2 string) (string, error)
	// RefExists checks if a git ref exists.
	RefExists(ref string) bool
	// CurrentBranch returns the current branch name.
	CurrentBranch() (string, error)
}

// ShellGitOps implements GitOps using shell commands via pkg/sh.
type ShellGitOps struct {
	Dir string // optional working directory (empty = current)
}

// NewGitOps creates a new ShellGitOps instance.
func NewGitOps(dir string) *ShellGitOps {
	return &ShellGitOps{Dir: dir}
}

// ChangedFiles returns files changed between base and head.
func (g *ShellGitOps) ChangedFiles(base, head string) ([]string, error) {
	args := []string{"diff", "--name-only", base, head}
	out, err := g.gitOutput(args...)
	if err != nil {
		return nil, err
	}

	out = strings.TrimSpace(out)
	if out == "" {
		return []string{}, nil
	}

	files := strings.Split(out, "\n")
	result := make([]string, 0, len(files))
	for _, f := range files {
		f = strings.TrimSpace(f)
		if f != "" {
			result = append(result, f)
		}
	}
	return result, nil
}

// MergeBase finds the merge base between two refs.
func (g *ShellGitOps) MergeBase(ref1, ref2 string) (string, error) {
	out, err := g.gitOutput("merge-base", ref1, ref2)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// RefExists checks if a git ref exists.
func (g *ShellGitOps) RefExists(ref string) bool {
	_, err := g.gitOutput("show-ref", "--verify", "--quiet", ref)
	return err == nil
}

// CurrentBranch returns the current branch name.
func (g *ShellGitOps) CurrentBranch() (string, error) {
	out, err := g.gitOutput("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// gitOutput runs a git command and returns its output.
func (g *ShellGitOps) gitOutput(args ...string) (string, error) {
	if g.Dir != "" {
		args = append([]string{"-C", g.Dir}, args...)
	}
	return sh.Output("git", args...)
}

// ContainsFile checks if a file is in the list of changed files.
func ContainsFile(files []string, target string) bool {
	for _, file := range files {
		if strings.Contains(file, target) {
			return true
		}
	}

	return false
}
