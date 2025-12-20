package st

import (
	"context"
	"fmt"
	"reflect"
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

	_, _, err = checkF(Foo.Bare, nil)
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

	hasContext, isNamespace, err = checkF(
		func(int, bool, string, time.Duration) {},
		[]interface{}{1, true, "s", time.Second},
	)
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
		if r := recover(); r != nil {
			t.Error("expected a nil function argument to be handled gracefully")
		}
	}()
	_, _, err = checkF(nil, []interface{}{1, 2})
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
	theFunc := func(cctx context.Context, ii int, ss string, bb bool, dd time.Duration) error {
		ctxOut = cctx //nolint:fatcontext // This is for the sake of the test.
		iOut = ii
		sOut = ss
		bOut = bb
		dOut = dd
		return nil
	}

	ctx := context.Background()
	iVal := 1776
	sVal := "abc124"
	bVal := true
	dVal := time.Second

	CtxDeps(ctx, F(theFunc, iVal, sVal, bVal, dVal))
	if GetCurrentTarget(ctxOut) == "" {
		t.Error("context missing target")
	}
	if iOut != iVal {
		t.Error(iOut)
	}
	if bOut != bVal {
		t.Error(bOut)
	}
	if dOut != dVal {
		t.Error(dOut)
	}
	if sOut != sVal {
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
		_, _ = fmt.Println(i)
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
	testFn := F(func(args ...string) {
		if !reflect.DeepEqual(args, []string{"a", "b"}) {
			t.Errorf("Wrong args, got %v, want [a b]", args)
		}
	}, "a", "b")
	err := testFn.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	//nolint:revive // Let's keep this as it is for the sake of the test.
	testFn = F(func(a string, b ...string) {}, "a", "b1", "b2")
	err = testFn.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	//nolint:revive // Let's keep this as it is for the sake of the test.
	testFn = F(func(a ...string) {})
	err = testFn.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	func() {
		defer func() {
			panicErr, ok := recover().(error)
			if !ok {
				t.Fatalf("expected panic with an error value, but got %T instead", recover())
			}
			wantMsg := "too few arguments for target, got 0"
			if panicErr == nil || panicErr.Error() != wantMsg {
				t.Fatal(panicErr)
			}
		}()
		//nolint:revive // Let's keep this as it is for the sake of the test.
		F(func(a string, b ...string) {})
	}()
}

type Foo Namespace

func (Foo) Bare() {}

func (Foo) Error() error { return nil }

func (Foo) BareCtx(context.Context) {}

func (Foo) CtxError(context.Context) error { return nil }

//nolint:revive // Let's keep this as it is for the sake of the test.
func (Foo) CtxErrorArgs(
	ctx context.Context, i int, s string, b bool, d time.Duration,
) error {
	return nil
}
