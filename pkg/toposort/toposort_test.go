package toposort

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type testNode struct {
	id   string
	deps []string
}

func (n testNode) TPID() string              { return n.id }
func (n testNode) DependencyTPIDs() []string { return n.deps }

func TestSort_SimpleChain(t *testing.T) {
	nodes := []testNode{
		{id: "A"},
		{id: "B", deps: []string{"A"}},
		{id: "C", deps: []string{"B"}},
	}

	out, err := Sort(nodes, false)
	require.NoError(t, err)
	require.Equal(t, []string{"A", "B", "C"}, ids(out))
}

func TestSort_MultipleRootsDeterministic(t *testing.T) {
	nodes := []testNode{
		{id: "A"},
		{id: "B"},
		{id: "C", deps: []string{"A", "B"}},
	}

	out, err := Sort(nodes, false)
	require.NoError(t, err)
	// Deterministic: A and B both roots, expect lexicographic order A, B, then C
	require.Equal(t, []string{"A", "B", "C"}, ids(out))
}

func TestSort_IndependentChains(t *testing.T) {
	nodes := []testNode{
		{id: "A"},
		{id: "B", deps: []string{"A"}},
		{id: "C"},
		{id: "D", deps: []string{"C"}},
	}

	out, err := Sort(nodes, false)
	require.NoError(t, err)
	// With lexicographic selection, we expect A, B, C, D
	require.Equal(t, []string{"A", "B", "C", "D"}, ids(out))
}

func TestSort_SelfCycle(t *testing.T) {
	nodes := []testNode{
		{id: "A", deps: []string{"A"}},
	}

	_, err := Sort(nodes, false)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrCircularDependency))
}

func TestSort_TwoNodeCycle(t *testing.T) {
	nodes := []testNode{
		{id: "A", deps: []string{"B"}},
		{id: "B", deps: []string{"A"}},
	}

	_, err := Sort(nodes, false)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrCircularDependency))
}

func TestSort_Empty(t *testing.T) {
	var nodes []testNode
	out, err := Sort(nodes, false)
	require.NoError(t, err)
	require.Len(t, out, 0)
}

// helper to extract ids
func ids[T interface{ TopoSortable }](items []T) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.TPID()
	}
	return out
}
