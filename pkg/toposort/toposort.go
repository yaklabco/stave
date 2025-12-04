package toposort

import (
	"errors"
	"fmt"
	"sort"
)

var (
	// ErrCircularDependency is returned when the input contains a dependency cycle.
	ErrCircularDependency = errors.New("circular dependency detected")

	// ErrMissingDependency is returned when a required dependency is not found.
	ErrMissingDependency = errors.New("dependency not found")
)

// Sort performs a topological sort of the provided items using Kahn's algorithm.
//
// The type parameter allows callers to receive back the same concrete element type
// as provided as long as it implements TopoSortable.
//
// If a cycle is detected the function returns an error that wraps ErrCircularDependency.
// The error message will include a list of nodes involved in the cycle (or remaining
// nodes that could not be resolved).
//
// If ignoreMissingDeps is true, dependencies that are not present in the provided
// items slice are ignored instead of causing an error.
func Sort[T interface{ TopoSortable }](items []T, ignoreMissingDeps bool) ([]T, error) {
	n := len(items)
	if n == 0 {
		return nil, nil
	}

	// Build ID -> item map and indegree counts based on dependencies.
	idToItem := make(map[string]T, n)
	for _, it := range items {
		id := it.TPID()
		idToItem[id] = it
	}

	// adjacency list: dep -> list of nodes that depend on it
	adj := make(map[string][]string, n)
	indeg := make(map[string]int, n)

	// initialize indegree for each node
	for id := range idToItem {
		indeg[id] = 0
	}

	// Build graph edges
	for _, it := range items {
		id := it.TPID()
		for _, dep := range it.DependencyTPIDs() {
			if dep == id {
				// self-cycle
				return nil, fmt.Errorf("%w: self dependency at %q", ErrCircularDependency, id)
			}
			// If dependency is unknown, consider it as an external node and error out
			// to avoid silently accepting incomplete graphs.
			if _, ok := idToItem[dep]; !ok {
				if ignoreMissingDeps {
					// Skip edges to missing nodes
					continue
				}
				return nil, fmt.Errorf("dependency %q of %q not found: %w", dep, id, ErrMissingDependency)
			}
			adj[dep] = append(adj[dep], id)
			indeg[id]++
		}
	}

	// Collect nodes with zero indegree and keep deterministic order.
	zeros := make([]string, 0, n)
	for id, d := range indeg {
		if d == 0 {
			zeros = append(zeros, id)
		}
	}
	sort.Strings(zeros)

	result := make([]T, 0, n)

	for len(zeros) > 0 {
		// pop the first (lexicographically smallest) for determinism
		id := zeros[0]
		zeros = zeros[1:]
		result = append(result, idToItem[id])
		for _, nei := range adj[id] {
			indeg[nei]--
			if indeg[nei] == 0 {
				// insert maintaining sorted order (simple append+sort ok for small N)
				zeros = append(zeros, nei)
			}
		}
		sort.Strings(zeros)
	}

	if len(result) != n {
		// find remaining nodes (those with indegree > 0)
		remaining := make([]string, 0, n-len(result))
		for id, d := range indeg {
			if d > 0 {
				remaining = append(remaining, id)
			}
		}
		sort.Strings(remaining)
		return nil, fmt.Errorf("%w: cycle among nodes: %v", ErrCircularDependency, remaining)
	}

	return result, nil
}
