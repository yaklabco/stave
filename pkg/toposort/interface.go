package toposort

type TopoSortable interface {
	TPID() string
	DependencyTPIDs() []string
}
