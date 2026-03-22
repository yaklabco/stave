//go:build stave

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Exits after receiving SIGHUP.
func ExitsAfterSighup(ctx context.Context) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGHUP)
	fmt.Println("ready")
	<-sigC
	fmt.Println("received sighup")
}

// Exits after SIGINT and wait.
func ExitsAfterSigint(ctx context.Context) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT)
	fmt.Println("ready")
	<-sigC
	fmt.Printf("exiting...")
	time.Sleep(200 * time.Millisecond)
	fmt.Println("done")
}

// Exits after ctx cancel and wait.
func ExitsAfterCancel(ctx context.Context) {
	defer func() {
		fmt.Println("deferred cleanup")
	}()
	fmt.Println("ready")
	<-ctx.Done()
	fmt.Printf("exiting...")
	time.Sleep(200 * time.Millisecond)
	fmt.Println("done")
}

// Ignores all signals, requires killing via timeout or second SIGINT.
func IgnoresSignals(ctx context.Context) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT)
	fmt.Println("ready")
	for {
		<-sigC
	}
}
