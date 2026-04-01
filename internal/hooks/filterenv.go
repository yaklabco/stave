package hooks

import "strings"

// FilterGitEnv removes GIT_DIR and GIT_WORK_TREE from the environment.
func FilterGitEnv(env []string) []string {
	var filtered []string
	for _, e := range env {
		if strings.HasPrefix(e, "GIT_DIR=") || strings.HasPrefix(e, "GIT_WORK_TREE=") {
			continue
		}
		filtered = append(filtered, e)
	}
	return filtered
}
