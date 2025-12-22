package st

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
)

// Fn represents a function that can be run with st.Deps. Package, Name, and ID must combine to
// uniquely identify a function, while ensuring the "same" function has identical values. These are
// used as a map key to find and run (or not run) the function.
type Fn interface {
	// Name should return the fully qualified name of the function. Usually
	// it's best to use runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name().
	Name() string

	// ID should be an additional uniqueness qualifier in case the name is insufficiently unique.
	// This can be the case for functions that take arguments (st.F json-encodes an array of the
	// args).
	ID() string

	// Run should run the function.
	Run(ctx context.Context) error

	// Underlying should return the original, wrapped function object.
	Underlying() *runtime.Func
}

// F takes a function that is compatible as a stave target, and any args that need to be passed to
// it, and wraps it in an st.Fn that st.Deps can run. Args must be passed in the same order as they
// are declared by the function. Note that you do not need to and should not pass a context.Context
// to F, even if the target takes a context. Compatible args are int, bool, string, and
// time.Duration.
func F(target any, args ...any) Fn {
	hasContext, isNamespace, err := checkF(target, args)
	if err != nil {
		panic(err)
	}
	argsID, err := json.Marshal(args)
	if err != nil {
		panic(fmt.Errorf("can't convert args into a stave-compatible id for st.Deps: %w", err))
	}
	return fn{
		name:       funcName(target),
		id:         string(argsID),
		f:          buildRunner(target, args, hasContext, isNamespace),
		underlying: funcObj(target),
	}
}

// buildRunner creates the runner function for a target with the given arguments.
func buildRunner(target any, args []any, hasContext, isNamespace bool) func(context.Context) error {
	return func(ctx context.Context) error {
		vargs := buildCallArgs(ctx, args, hasContext, isNamespace)
		return callAndHandleResult(reflect.ValueOf(target), vargs)
	}
}

// buildCallArgs constructs the reflect.Value slice for calling the target function.
func buildCallArgs(ctx context.Context, args []any, hasContext, isNamespace bool) []reflect.Value {
	count := len(args)
	if hasContext {
		count++
	}
	if isNamespace {
		count++
	}
	vargs := make([]reflect.Value, count)
	argIndex := 0
	if isNamespace {
		vargs[0] = reflect.ValueOf(struct{}{})
		argIndex++
	}
	if hasContext {
		vargs[argIndex] = reflect.ValueOf(ctx)
		argIndex++
	}
	for idx := range args {
		vargs[argIndex+idx] = reflect.ValueOf(args[idx])
	}
	return vargs
}

// callAndHandleResult calls the function and handles the error return value.
func callAndHandleResult(theValue reflect.Value, vargs []reflect.Value) error {
	ret := theValue.Call(vargs)
	if len(ret) == 0 {
		return nil
	}
	// we only allow functions with a single error return, so this should be safe.
	if ret[0].IsNil() {
		return nil
	}
	retErr, ok := ret[0].Interface().(error)
	if !ok {
		return fmt.Errorf("expected function to return an error, but got %T instead", ret[0].Interface())
	}
	return retErr
}

type fn struct {
	name       string
	id         string
	f          func(ctx context.Context) error
	underlying *runtime.Func
}

// Name returns the fully qualified name of the function.
func (f fn) Name() string {
	return f.name
}

// ID returns a hash of the argument values passed in.
func (f fn) ID() string {
	return f.id
}

// Run runs the function.
func (f fn) Run(ctx context.Context) error {
	return f.f(ctx)
}

// Underlying returns the original, wrapped function object.
func (f fn) Underlying() *runtime.Func {
	return f.underlying
}

func checkF(target any, args []any) (bool, bool, error) {
	theType := reflect.TypeOf(target)
	if err := validateTargetType(theType, target); err != nil {
		return false, false, err
	}
	if err := validateReturnType(theType); err != nil {
		return false, false, err
	}
	if err := validateArgCount(theType, args); err != nil {
		return false, false, err
	}
	if theType.NumIn() == 0 {
		return false, false, nil
	}
	return validateArgs(theType, args)
}

// validateTargetType checks that target is a function.
func validateTargetType(theType reflect.Type, target any) error {
	if theType == nil || theType.Kind() != reflect.Func {
		return fmt.Errorf(
			"non-function passed to st.F: %T. "+
				"The st.F function accepts function names, such as st.F(TargetA, \"arg1\", \"arg2\")",
			target,
		)
	}
	return nil
}

// validateReturnType checks the function has zero or one error return.
func validateReturnType(theType reflect.Type) error {
	if theType.NumOut() > 1 {
		return errors.New("target has too many return values, must be zero or just an error")
	}
	if theType.NumOut() == 1 && theType.Out(0) != errType {
		return errors.New("target's return value is not an error")
	}
	return nil
}

// validateArgCount checks the number of arguments is valid for the target.
func validateArgCount(theType reflect.Type, args []any) error {
	if len(args) > theType.NumIn() && !theType.IsVariadic() {
		return fmt.Errorf("too many arguments for target, got %d", len(args))
	}
	return nil
}

// validateArgs checks each argument matches the expected type and returns context/namespace flags.
func validateArgs(theType reflect.Type, args []any) (bool, bool, error) {
	argIndex := 0
	inputs := theType.NumIn()
	isNamespace := false
	hasContext := false

	// Check for namespace receiver
	if theType.In(0).AssignableTo(emptyType) {
		isNamespace = true
		argIndex++
		inputs-- // callers must leave off the namespace value
	}

	// Check for context parameter
	if theType.NumIn() > argIndex && theType.In(argIndex) == ctxType {
		inputs-- // callers must leave off the context
		hasContext = true
		argIndex++ // skip context in argument checking loop
	}

	// Validate argument count
	if err := checkArgCountForVariadic(theType, args, inputs); err != nil {
		return false, false, err
	}

	// Validate each argument type
	if err := checkArgTypes(theType, args, argIndex); err != nil {
		return false, false, err
	}

	return hasContext, isNamespace, nil
}

// checkArgCountForVariadic validates argument count considering variadic functions.
func checkArgCountForVariadic(theType reflect.Type, args []any, inputs int) error {
	if theType.IsVariadic() {
		if len(args) < inputs-1 {
			return fmt.Errorf("too few arguments for target, got %d", len(args))
		}
	} else if len(args) != inputs {
		return fmt.Errorf("wrong number of arguments for target, got %d", len(args))
	}
	return nil
}

// checkArgTypes validates each argument matches its expected type.
func checkArgTypes(theType reflect.Type, args []any, startIndex int) error {
	argIndex := startIndex
	for _, arg := range args {
		argType := theType.In(argIndex)
		if theType.IsVariadic() && argIndex == theType.NumIn()-1 {
			argType = argType.Elem() // For variadic, use the slice element type
		}
		if !argTypes[argType] {
			return fmt.Errorf("argument %d (%s), is not a supported argument type", argIndex, argType)
		}
		passedType := reflect.TypeOf(arg)
		if argType != passedType {
			return fmt.Errorf("argument %d expected to be %s, but is %s", argIndex, argType, passedType)
		}
		if argIndex < theType.NumIn()-1 {
			argIndex++
		}
	}
	return nil
}
