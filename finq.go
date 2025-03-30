package finq

import (
	"context"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type MonitorOpts struct {
	// PauseBetweenChecks is the pause duration between each check iteration.
	PauseBetweenChecks time.Duration

	// StallingInterval is the interval at which OnStalling will be called for a stalling finalizer.
	// StallingInterval must be > 0.
	StallingInterval time.Duration

	// OnStalling is called when the finalizer is stalling every StallingInterval.
	// OnStalling is mandatory.
	OnStalling func(d time.Duration)

	// OnComplete is called when the finalizer is executed.
	OnComplete func(d time.Duration)

	// ImmediatelyTriggerGC will trigger a GC at the beginning of each check iteration.
	ImmediatelyTriggerGC bool
}

func Monitor(ctx context.Context, opts *MonitorOpts) {
	if opts == nil {
		panic("nil MonitorOpts")
	}
	if opts.OnStalling == nil {
		panic("nil OnStalling")
	}
	if opts.StallingInterval <= 0 {
		panic("invalid StallingInterval")
	}

	for {
		checkFinalizerOnce(ctx, opts)
		select {
		case <-ctx.Done():
			return
		case <-time.After(opts.PauseBetweenChecks):
		}
	}
}

// checkFinalizerOnce will create an object with a finalizer and wait for the finalizer to be executed.
// while waiting, it will call onUpdate with the time since the start of the function.
func checkFinalizerOnce(ctx context.Context, opts *MonitorOpts) {
	finalizerCtx, finalizerCancel := context.WithCancel(context.Background())
	defer finalizerCancel()

	start := time.Now()
	createAndDeleteObjectWithFinalizer(finalizerCancel)

	if opts.ImmediatelyTriggerGC {
		runtime.GC()
	}

	ticker := time.NewTicker(opts.StallingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-finalizerCtx.Done():
			if opts.OnComplete != nil {
				opts.OnComplete(time.Now().Sub(start))
			}
			return
		case now := <-ticker.C:
			opts.OnStalling(now.Sub(start))
		}
	}
}

// createAndDeleteObjectWithFinalizer creates an object with a finalizer that will execute the finalizerFunc.
func createAndDeleteObjectWithFinalizer(finalizerFunc func()) {
	// create a dummy object
	x := new(*int)

	// set a finalizer to cancel the context
	runtime.SetFinalizer(x, func(_ **int) {
		finalizerFunc()
	})

	// make sure x is not optimized away by the compiler
	runtime.KeepAlive(x)

	// remove the reference to x so it will be unreachable and its finalizer will be triggered
	x = nil
}

// GetFinalizerStackTrace attempts to return the stack trace of the runtime goroutine that is running the finalizer.
func GetFinalizerStackTrace() string {
	stacks := getGoRoutinesStacks(runtime.GoroutineProfile, true)
	for _, s := range stacks {
		if strings.Contains(s, "runtime.runfinq") {
			return s
		}
	}
	return ""
}

// getGoRoutinesStacks is a modified copy of 'runtime/pprof' pkg test func getProfileStacks.
// it returns the stack traces of all goroutines.
func getGoRoutinesStacks(collect func([]runtime.StackRecord) (int, bool), fileLine bool) []string {
	var n int
	var ok bool
	var p []runtime.StackRecord
	for {
		p = make([]runtime.StackRecord, n)
		n, ok = collect(p)
		if ok {
			p = p[:n]
			break
		}
	}
	var stacks []string
	for _, r := range p {
		var stack strings.Builder
		for i, pc := range r.Stack() {
			if i > 0 {
				stack.WriteByte('\n')
			}
			// Use FuncForPC instead of CallersFrames,
			// because we want to see the info for exactly
			// the PCs returned by the mutex profile to
			// ensure inlined calls have already been properly
			// expanded.
			f := runtime.FuncForPC(pc - 1)
			stack.WriteString(f.Name())
			if fileLine {
				stack.WriteByte(' ')
				file, line := f.FileLine(pc - 1)
				stack.WriteString(file)
				stack.WriteByte(':')
				stack.WriteString(strconv.Itoa(line))
			}
		}
		stacks = append(stacks, stack.String())
	}
	return stacks
}
