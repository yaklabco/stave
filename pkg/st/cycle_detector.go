package st

import (
	"log/slog"
	"runtime"
	"strings"
	"sync"

	"github.com/samber/lo"
	"github.com/yaklabco/stave/pkg/toposort"
)

const (
	// maxStackDepthToCheck defines the maximum stack depth for runtime caller inspection.
	maxStackDepthToCheck = 64
)

var (
	depsByID      = make(map[string]toposort.TopoSortable) //nolint:gochecknoglobals // Part of a mutexed pattern.
	depsByIDMutex sync.RWMutex                             //nolint:gochecknoglobals // Part of a mutexed pattern.
)

func firstExternalCaller() *runtime.Frame {
	thisProgCtr, _, _, ok := runtime.Caller(0)
	if !ok {
		return nil
	}
	thisFunc := runtime.FuncForPC(thisProgCtr)
	pkgPrefix := getPackagePath(thisFunc)

	// runtime.Callers (0), firstExternalCaller (1), the function calling firstExternalCaller (2)
	const skip = 2
	progCtrsAboveUs := make([]uintptr, maxStackDepthToCheck)
	nProgCtrsAboveUs := runtime.Callers(skip, progCtrsAboveUs)
	frames := runtime.CallersFrames(progCtrsAboveUs[:nProgCtrsAboveUs])

	for {
		frame, more := frames.Next()

		// frame.Function is the fully-qualified name:
		//   "mypkg.myFunc"
		//   "otherpkg.DoThing"
		//   "github.com/me/foo/bar.Baz"
		slog.Debug("checking the frame", slog.String("function", frame.Function), slog.String("pkg_prefix", pkgPrefix))
		if !strings.HasPrefix(frame.Function, pkgPrefix) {
			return &frame
		}

		if !more {
			break
		}
	}

	return nil
}

func getPackagePath(thisFunc *runtime.Func) string {
	pkgPrefix := thisFunc.Name()
	lastSlash := strings.LastIndex(pkgPrefix, "/")
	lastDot := strings.LastIndex(pkgPrefix, ".")
	if lastDot > lastSlash {
		pkgPrefix = pkgPrefix[:lastDot] // e.g. "github.com/me/project/mypkg"
	}
	return pkgPrefix
}

func checkForCycle(funcs []Fn) error {
	callerFrame := firstExternalCaller()
	if callerFrame == nil {
		slog.Warn("could not determine caller, skipping circular-dependency check")
		return nil
	}

	callerID := callerFrame.File + ":" + callerFrame.Function
	slog.Debug("checking for cycle", slog.String("caller_id", callerID))

	funcIDs := make([]string, 0, len(funcs))
	for _, theFunc := range funcs {
		theFuncObj := theFunc.Underlying()
		theFile, _ := theFuncObj.FileLine(0)
		theFuncID := theFile + ":" + theFuncObj.Name()
		slog.Debug("adding dependency", slog.String("func_id", theFuncID))
		funcIDs = append(funcIDs, theFuncID)
	}

	depsByIDMutex.Lock()
	defer depsByIDMutex.Unlock()
	depsByID[callerID] = depsNode{tpID: callerID, dependencyTPIDs: funcIDs}

	_, err := toposort.Sort(lo.Values(depsByID), true)

	return err
}

type depsNode struct {
	tpID            string
	dependencyTPIDs []string
}

func (n depsNode) TPID() string              { return n.tpID }
func (n depsNode) DependencyTPIDs() []string { return n.dependencyTPIDs }
