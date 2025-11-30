package mg

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestFuncCheck(t *testing.T) {
	hasContext, isNamespace, err := checkF(func() {}, nil)
	if err != nil {
		t.Error(err)
	}
	if hasContext {
		t.Error("func does not have context")
	}
	if isNamespace {
		t.Error("func is not on a namespace")
	}
	hasContext, isNamespace, err = checkF(func() error { return nil }, nil)
	if err != nil {
		t.Error(err)
	}
	if hasContext {
		t.Error("func does not have context")
	}
	if isNamespace {
		t.Error("func is not on a namespace")
	}
	hasContext, isNamespace, err = checkF(func(context.Context) {}, nil)
	if err != nil {
		t.Error(err)
	}
	if !hasContext {
		t.Error("func has context")
	}
	if isNamespace {
		t.Error("func is not on a namespace")
	}
	hasContext, isNamespace, err = checkF(func(context.Context) error { return nil }, nil)
	if err != nil {
		t.Error(err)
	}
	if !hasContext {
		t.Error("func has context")
	}
	if isNamespace {
		t.Error("func is not on a namespace")
	}

	hasContext, isNamespace, err = checkF(Foo.Bare, nil)
	if err != nil {
		t.Error(err)
	}

	hasContext, isNamespace, err = checkF(Foo.Error, nil)
	if err != nil {
		t.Error(err)
	}
	if hasContext {
		t.Error("func does not have context")
	}
	if !isNamespace {
		t.Error("func is  on a namespace")
	}

	hasContext, isNamespace, err = checkF(Foo.BareCtx, nil)
	if err != nil {
		t.Error(err)
	}
	if !hasContext {
		t.Error("func has context")
	}
	if !isNamespace {
		t.Error("func is  on a namespace")
	}
	hasContext, isNamespace, err = checkF(Foo.CtxError, nil)
	if err != nil {
		t.Error(err)
	}
	if !hasContext {
		t.Error("func has context")
	}
	if !isNamespace {
		t.Error("func is  on a namespace")
	}

	hasContext, isNamespace, err = checkF(Foo.CtxErrorArgs, []interface{}{1, "s", true, time.Second})
	if err != nil {
		t.Error(err)
	}
	if !hasContext {
		t.Error("func has context")
	}
	if !isNamespace {
		t.Error("func is on a namespace")
	}

	hasContext, isNamespace, err = checkF(func(int, bool, string, time.Duration) {}, []interface{}{1, true, "s", time.Second})
	if err != nil {
		t.Error(err)
	}
	if hasContext {
		t.Error("func does not have context")
	}
	if isNamespace {
		t.Error("func is not on a namespace")
	}

	// Test an Invalid case
	_, _, err = checkF(func(*int) error { return nil }, nil)
	if err == nil {
		t.Error("expected func(*int) error to be invalid")
	}

	defer func() {
		if r := recover(); r !=nil {
			t.Error("expected a nil function argument to be handled gracefully")
		}
	}()
	_, _, err = checkF(nil, []interface{}{1,2})
	if err == nil {
		t.Error("expected a nil function argument to be invalid")
	}
}

func TestF(t *testing.T) {
	var (
		ctxOut context.Context
		iOut   int
		sOut   string
		bOut   bool
		dOut   time.Duration
	)
	f := func(cctx context.Context, ii int, ss string, bb bool, dd time.Duration) error {
		ctxOut = cctx
		iOut = ii
		sOut = ss
		bOut = bb
		dOut = dd
		return nil
	}

	ctx := context.Background()
	i := 1776
	s := "abc124"
	b := true
	d := time.Second

	CtxDeps(ctx, F(f, i, s, b, d))
	if ctxOut != ctx {
		t.Error(ctxOut)
	}
	if iOut != i {
		t.Error(iOut)
	}
	if bOut != b {
		t.Error(bOut)
	}
	if dOut != d {
		t.Error(dOut)
	}
	if sOut != s {
		t.Error(sOut)
	}
}

func TestFTwice(t *testing.T) {
	var called int64
	f := func(int) {
		atomic.AddInt64(&called, 1)
	}

	Deps(F(f, 5), F(f, 5), F(f, 1))
	if called != 2 {
		t.Fatalf("Expected to be called 2 times, but was called %d", called)
	}
}

func ExampleF() {
	f := func(i int) {
		fmt.Println(i)
	}

	// we use SerialDeps here to ensure consistent output, but this works with all Deps functions.
	SerialDeps(F(f, 5), F(f, 1))
	// output:
	// 5
	// 1
}

func TestFNamespace(t *testing.T) {
	ctx := context.Background()
	i := 1776
	s := "abc124"
	b := true
	d := time.Second

	fn := F(Foo.CtxErrorArgs, i, s, b, d)
	err := fn.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFNilError(t *testing.T) {
	fn := F(func() error { return nil })
	err := fn.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestFVariadic(t *testing.T) {
	fn := F(func(args ...string) {
		if !reflect.DeepEqual(args, []string{"a", "b"}) {
			t.Errorf("Wrong args, got %v, want [a b]", args)
		}
	}, "a", "b")
	err := fn.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	fn = F(func(a string, b ...string) {}, "a", "b1", "b2")
	err = fn.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	fn = F(func(a ...string) {})
	err = fn.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	func() {
		defer func() {
			err, _ := recover().(error)
			if err == nil || err.Error() != "too few arguments for target, got 0 for func(string, ...string)" {
				t.Fatal(err)
			}
		}()
		F(func(a string, b ...string) {})
	}()
}

func TestFnNameAndID(t *testing.T) {
	t.Parallel()

	testFunc := func(i int, s string) {}
	fn := F(testFunc, 42, "test")

	// Verify Name() returns the function name
	name := fn.Name()
	if name == "" {
		t.Error("fn.Name() returned empty string")
	}
	if !strings.Contains(name, "mg") {
		t.Errorf("fn.Name() = %q, expected to contain 'mg'", name)
	}

	// Verify ID() returns the JSON-encoded args
	id := fn.ID()
	if !strings.Contains(id, "42") || !strings.Contains(id, "test") {
		t.Errorf("fn.ID() = %q, expected to contain '42' and 'test'", id)
	}

	// Verify different args produce different IDs
	fn2 := F(testFunc, 99, "different")
	if fn.ID() == fn2.ID() {
		t.Error("fn.ID() should be different for different args")
	}

	// Verify same args produce same ID
	fn3 := F(testFunc, 42, "test")
	if fn.ID() != fn3.ID() {
		t.Errorf("fn.ID() should be same for same args: %q != %q", fn.ID(), fn3.ID())
	}
}

func TestFPanicOnNonFunction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		target interface{}
	}{
		{"nil", nil},
		{"int", 42},
		{"string", "not a function"},
		{"struct", struct{}{}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				if r := recover(); r == nil {
					t.Error("F() should panic on non-function, but didn't")
				}
			}()

			F(tt.target)
		})
	}
}

func TestFPanicOnBadArgs(t *testing.T) {
	t.Parallel()

	// Channels are not valid argument types
	ch := make(chan int)

	defer func() {
		if r := recover(); r == nil {
			t.Error("F() should panic on invalid argument types, but didn't")
		} else {
			errMsg := fmt.Sprint(r)
			// The error could be about type mismatch or JSON marshaling
			if !strings.Contains(errMsg, "expected to be int") && !strings.Contains(errMsg, "json") && !strings.Contains(errMsg, "JSON") {
				t.Errorf("expected error about type or JSON, got: %v", r)
			}
		}
	}()

	testFunc := func(i int) {}
	F(testFunc, ch) // This should panic because chan is not a valid argument type
}

func TestCheckFErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		target    interface{}
		args      []interface{}
		wantPanic bool
	}{
		{
			name:      "too many returns",
			target:    func() (int, error) { return 0, nil },
			wantPanic: true,
		},
		{
			name:      "bad return type",
			target:    func() int { return 0 },
			wantPanic: true,
		},
		{
			name:      "too many args",
			target:    func(i int) {},
			args:      []interface{}{1, 2},
			wantPanic: true,
		},
		{
			name:      "wrong arg type",
			target:    func(i int) {},
			args:      []interface{}{"not an int"},
			wantPanic: true,
		},
		{
			name:      "valid function",
			target:    func(i int) error { return nil },
			args:      []interface{}{42},
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("F() panic = %v, wantPanic %v", r != nil, tt.wantPanic)
				}
			}()

			F(tt.target, tt.args...)
		})
	}
}

type Foo Namespace

func (Foo) Bare() {}

func (Foo) Error() error { return nil }

func (Foo) BareCtx(context.Context) {}

func (Foo) CtxError(context.Context) error { return nil }

func (Foo) CtxErrorArgs(ctx context.Context, i int, s string, b bool, d time.Duration) error {
	return nil
}
