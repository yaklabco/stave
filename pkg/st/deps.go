package st

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/yaklabco/stave/internal/log"
	"github.com/yaklabco/stave/pkg/watch/target/wctx"
)

type onceMap struct {
	mu *sync.Mutex
	m  map[onceKey]*onceFun
}

// ContextWithTarget returns a new context with the target name attached.
func ContextWithTarget(ctx context.Context, name string) context.Context {
	return wctx.ContextWithTarget(ctx, name)
}

// GetCurrentTarget returns the target name from the context, or empty string if not found.
func GetCurrentTarget(ctx context.Context) string {
	return wctx.GetCurrentTarget(ctx)
}

// ContextWithTargetState returns a new context with the target state attached.
func ContextWithTargetState(ctx context.Context, state any) context.Context {
	return wctx.ContextWithTargetState(ctx, state)
}

// GetTargetState returns the target state from the context, or nil if not found.
func GetTargetState(ctx context.Context) any {
	return wctx.GetTargetState(ctx)
}

// SetOverallWatchMode sets whether we are in overall watch mode.
func SetOverallWatchMode(b bool) {
	wctx.SetOverallWatchMode(b)
}

// IsOverallWatchMode returns whether we are in overall watch mode.
func IsOverallWatchMode() bool {
	return wctx.IsOverallWatchMode()
}

// SetOutermostTarget sets the name of the outermost target.
func SetOutermostTarget(name string) {
	wctx.SetOutermostTarget(name)
}

// GetOutermostTarget returns the name of the outermost target.
func GetOutermostTarget() string {
	return wctx.GetOutermostTarget()
}

type onceKey struct {
	Name string
	ID   string
}

func (o *onceMap) LoadOrStore(theFunc Fn) *onceFun {
	defer o.mu.Unlock()
	o.mu.Lock()

	key := onceKey{
		Name: theFunc.Name(),
		ID:   theFunc.ID(),
	}
	existing, ok := o.m[key]
	if ok {
		return existing
	}
	one := &onceFun{
		once:        &sync.Once{},
		fn:          theFunc,
		displayName: DisplayName(theFunc.Name()),
	}
	o.m[key] = one
	return one
}

// SerialDeps is like Deps except it runs each dependency serially, instead of
// in parallel. This can be useful for resource intensive dependencies that
// shouldn't be run at the same time.
func SerialDeps(fns ...any) {
	funcs := checkFns(fns)
	ctx := wctx.GetActiveContext()
	for i := range fns {
		runDeps(ctx, funcs[i:i+1])
	}
}

// SerialCtxDeps is like CtxDeps except it runs each dependency serially,
// instead of in parallel. This can be useful for resource intensive
// dependencies that shouldn't be run at the same time.
func SerialCtxDeps(ctx context.Context, fns ...any) {
	funcs := checkFns(fns)
	for i := range fns {
		runDeps(ctx, funcs[i:i+1])
	}
}

// CtxDeps runs the given functions as dependencies of the calling function.
// Dependencies must only be of type:
//
//	func()
//	error
//	func(context.Context)
//	error
//
// Or a similar method on a st.Namespace type.
// Or an st.Fn interface.
//
// The function calling Deps is guaranteed that all dependent functions will be
// run exactly once when Deps returns.  Dependent functions may in turn declare
// their own dependencies using Deps. Each dependency is run in their own
// goroutines. Each function is given the context provided if the function
// prototype allows for it.
func CtxDeps(ctx context.Context, fns ...any) {
	funcs := checkFns(fns)
	runDeps(ctx, funcs)
}

// runDeps assumes you've already called checkFns.
func runDeps(ctx context.Context, fns []Fn) {
	errMutex := &sync.Mutex{}
	var errs []string
	var exit int
	waitGroup := &sync.WaitGroup{}
	for _, depFn := range fns {
		depFunc := onces.LoadOrStore(depFn)
		waitGroup.Add(1)
		go func() {
			defer func() {
				if panicValue := recover(); panicValue != nil {
					errMutex.Lock()
					if err, ok := panicValue.(error); ok {
						exit = changeExit(exit, ExitStatus(err))
					} else {
						exit = changeExit(exit, 1)
					}
					errs = append(errs, fmt.Sprint(panicValue))
					errMutex.Unlock()
				}
				waitGroup.Done()
			}()
			if err := depFunc.run(ctx); err != nil {
				errMutex.Lock()
				errs = append(errs, fmt.Sprint(err))
				exit = changeExit(exit, ExitStatus(err))
				errMutex.Unlock()
			}
		}()
	}

	waitGroup.Wait()
	if len(errs) > 0 {
		panic(Fatal(exit, strings.Join(errs, "\n")))
	}
}

func checkFns(fns []any) []Fn {
	funcs := make([]Fn, len(fns))
	for iFunc, theFunc := range fns {
		if fn, ok := theFunc.(Fn); ok {
			funcs[iFunc] = fn
			continue
		}

		// Check if the target provided is a not function so we can give a clear warning
		t := reflect.TypeOf(theFunc)
		if t == nil || t.Kind() != reflect.Func {
			panic(fmt.Errorf("non-function used as a target dependency: %T. The st.Deps, st.SerialDeps and st.CtxDeps functions accept function names, such as st.Deps(TargetA, TargetB)", theFunc)) //nolint:lll // Long string-literal.
		}

		funcs[iFunc] = F(theFunc)
	}

	if err := checkForCycle(funcs); err != nil {
		panic(fmt.Errorf("checking for cycles in dependency graph: %w", err))
	}

	return funcs
}

// Deps runs the given functions in parallel, exactly once. Dependencies must
// only be of type:
//
//	func()
//	error
//	func(context.Context)
//	error
//
// Or a similar method on a st.Namespace type.
// Or an st.Fn interface.
//
// This is a way to build up a tree of dependencies with each dependency
// defining its own dependencies.  Functions must have the same signature as a
// Stave target, i.e. optional context argument, optional error return.
func Deps(fns ...any) {
	CtxDeps(wctx.GetActiveContext(), fns...)
}

func changeExit(oldExitCode, newExitCode int) int {
	if newExitCode == 0 {
		return oldExitCode
	}
	if oldExitCode == 0 {
		return newExitCode
	}
	if oldExitCode == newExitCode {
		return oldExitCode
	}
	// both different and both non-zero, just set
	// exit to 1. Nothing more we can do.
	return 1
}

// funcName returns the unique name for the function.
func funcName(i any) string {
	return funcObj(i).Name()
}

func funcObj(i any) *runtime.Func {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer())
}

func DisplayName(name string) string {
	return wctx.DisplayName(name)
}

type onceFun struct {
	once *sync.Once
	fn   Fn
	err  error

	displayName string
}

// run will run the function exactly once and capture the error output. Further runs simply return
// the same error output.
func (o *onceFun) run(ctx context.Context) error {
	ctx = ContextWithTarget(ctx, o.displayName)
	wctx.RegisterTargetContext(ctx, o.displayName)
	defer wctx.UnregisterTargetContext(o.displayName)
	o.once.Do(func() {
		if Verbose() {
			log.SimpleConsoleLogger.Println("Running dependency:", DisplayName(o.fn.Name()))
		}
		o.err = o.fn.Run(ctx)
	})
	return o.err
}

// RunFn runs the given function as a Stave target.
func RunFn(ctx context.Context, theFunc any) error {
	var fn Fn
	if f, ok := theFunc.(Fn); ok {
		fn = f
	} else {
		fn = F(theFunc)
	}
	displayName := DisplayName(fn.Name())
	ctx = ContextWithTarget(ctx, displayName)
	wctx.RegisterTargetContext(ctx, displayName)
	defer wctx.UnregisterTargetContext(displayName)
	return fn.Run(ctx)
}
