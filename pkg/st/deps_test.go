package st

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestDepsRunOnce(t *testing.T) {
	done := make(chan struct{})
	f := func() {
		done <- struct{}{}
	}
	go Deps(f, f)
	select {
	case <-done:
		// cool
	case <-time.After(time.Millisecond * 100):
		t.Fatal("func not run in a reasonable amount of time.")
	}
	select {
	case <-done:
		t.Fatal("func run twice!")
	case <-time.After(time.Millisecond * 100):
		// cool... this should be plenty of time for the goroutine to have run
	}
}

func TestDepsOfDeps(t *testing.T) {
	resultChan := make(chan string, 3)
	// this->f->g->h
	funcH := func() {
		resultChan <- "h"
	}
	funcG := func() {
		Deps(funcH)
		resultChan <- "g"
	}
	funcF := func() {
		Deps(funcG)
		resultChan <- "f"
	}
	Deps(funcF)

	res := <-resultChan + <-resultChan + <-resultChan

	if res != "hgf" {
		t.Fatal("expected h then g then f to run, but got " + res)
	}
}

func TestSerialDeps(t *testing.T) {
	resultChan := make(chan string, 3)
	// this->funcF->funcG->funcH
	funcH := func() {
		resultChan <- "h"
	}
	funcG := func() {
		resultChan <- "g"
	}
	funcF := func() {
		SerialDeps(funcG, funcH)
		resultChan <- "f"
	}
	Deps(funcF)

	res := <-resultChan + <-resultChan + <-resultChan

	if res != "ghf" {
		t.Fatal("expected funcG then funcH then funcF to run, but got " + res)
	}
}

func TestDepError(t *testing.T) {
	// TODO: this test is ugly and relies on implementation details. It should
	// be recreated as a full-stack test.

	theFunc := func() error {
		return errors.New("ouch")
	}
	defer func() {
		err := recover()
		if err == nil {
			t.Fatal("expected panic, but didn't get one")
		}
		actual := fmt.Sprint(err)
		if actual != "ouch" {
			t.Fatalf(`expected to get "ouch" but got "%s"`, actual)
		}
	}()
	Deps(theFunc)
}

func TestDepFatal(t *testing.T) {
	theFunc := func() error {
		return Fatal(99, "ouch!")
	}
	defer func() {
		panicValue := recover()
		if panicValue == nil {
			t.Fatal("expected panic, but didn't get one")
		}
		actual := fmt.Sprint(panicValue)
		if actual != "ouch!" {
			t.Fatalf(`expected to get "ouch!" but got "%s"`, actual)
		}
		err, ok := panicValue.(error)
		if !ok {
			t.Fatalf("expected recovered val to be error but was %T", panicValue)
		}
		code := ExitStatus(err)
		if code != 99 {
			t.Fatalf("Expected exit status 99, but got %v", code)
		}
	}()
	Deps(theFunc)
}

func TestDepTwoFatal(t *testing.T) {
	funcF := func() error {
		return Fatal(99, "ouch!")
	}
	funcG := func() error {
		return Fatal(11, "bang!")
	}
	defer func() {
		panicValue := recover()
		if panicValue == nil {
			t.Fatal("expected panic, but didn't get one")
		}
		actual := fmt.Sprint(panicValue)
		// order is non-deterministic, so check for both orders
		if actual != "ouch!\nbang!" && actual != "bang!\nouch!" {
			t.Fatalf(`expected to get "ouch!" and "bang!" but got "%s"`, actual)
		}
		err, ok := panicValue.(error)
		if !ok {
			t.Fatalf("expected recovered val to be error but was %T", panicValue)
		}
		code := ExitStatus(err)
		// two different error codes returns, so we give up and just use error
		// code 1.
		if code != 1 {
			t.Fatalf("Expected exit status 1, but got %v", code)
		}
	}()
	Deps(funcF, funcG)
}

func TestDepWithUnhandledFunc(t *testing.T) {
	defer func() {
		err := recover()
		_, ok := err.(error)
		if !ok {
			t.Fatalf("Expected type error from panic")
		}
	}()
	// notValidFunc has wrong signature (returns string, not error) to test panic handling
	notValidFunc := func(a string) string {
		return a
	}
	Deps(notValidFunc)
}

func TestDepsErrors(t *testing.T) {
	var hRan, gRan, fRan int64

	funcH := func() error {
		atomic.AddInt64(&hRan, 1)
		return errors.New("oops")
	}
	funcG := func() {
		Deps(funcH)
		atomic.AddInt64(&gRan, 1)
	}
	funcF := func() {
		Deps(funcG, funcH)
		atomic.AddInt64(&fRan, 1)
	}

	defer func() {
		err := recover()
		if err == nil {
			t.Fatal("expected funcF to panic")
		}
		if hRan != 1 {
			t.Fatalf("expected funcH to run once, but got %v", hRan)
		}
		if gRan > 0 {
			t.Fatalf("expected funcG to panic before incrementing gRan to run, but got %v", gRan)
		}
		if fRan > 0 {
			t.Fatalf("expected funcF to panic before incrementing fRan to run, but got %v", fRan)
		}
	}()
	funcF()
}
