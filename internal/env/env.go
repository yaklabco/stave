package env

import (
	"os"
	"strings"

	"github.com/samber/lo"
)

func GetMap() map[string]string {
	return ToMap(os.Environ())
}

func ToMap(assignments []string) map[string]string {
	return lo.FromPairs(lo.FilterMap(assignments, func(item string, _ int) (lo.Entry[string, string], bool) {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			return lo.Entry[string, string]{}, false
		}

		return lo.Entry[string, string]{Key: parts[0], Value: parts[1]}, true
	}))
}

func ToAssignments(envMap map[string]string) []string {
	return lo.MapToSlice(envMap, func(k, v string) string {
		return k + "=" + v
	})
}
