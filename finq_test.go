package finq

import (
	"context"
	"github.com/stretchr/testify/assert"
	"reflect"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func StuckFinalizerForTestingPurpose() context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())

	createAndDeleteObjectWithFinalizer(func() {
		<-ctx.Done()
	})

	runtime.GC()
	runtime.GC()

	return cancel
}

func TestCheckFinalizerOnce_Valid(t *testing.T) {
	calledOnComplete := make(chan struct{})
	calledOnStalling := atomic.Int32{}

	go checkFinalizerOnce(context.Background(), &MonitorOpts{
		StallingInterval: time.Millisecond,
		OnComplete: func(d time.Duration) {
			close(calledOnComplete)
		},
		OnStalling: func(d time.Duration) {
			calledOnStalling.Add(1)
			runtime.GC()
		},
		ImmediatelyTriggerGC: true,
	})

	select {
	case <-calledOnComplete:
	case <-time.After(time.Second):
		t.Error("timeout")
	}
}

func TestCheckFinalizerOnce_Stuck(t *testing.T) {
	cancel := StuckFinalizerForTestingPurpose()
	defer cancel()

	calledOnComplete := make(chan struct{})
	calledOnStalling := atomic.Int32{}

	go checkFinalizerOnce(context.Background(), &MonitorOpts{
		StallingInterval: time.Millisecond,
		OnComplete: func(d time.Duration) {
			close(calledOnComplete)
		},
		OnStalling: func(d time.Duration) {
			calledOnStalling.Add(1)
			runtime.GC()
		},
		ImmediatelyTriggerGC: true,
	})

	select {
	case <-calledOnComplete:
		t.Error("should not be executed")
	case <-time.After(time.Second):
		assert.NotZero(t, calledOnStalling.Load())
	}

	st := GetFinalizerStackTrace()
	stuckFuncName := runtime.FuncForPC(reflect.ValueOf(StuckFinalizerForTestingPurpose).Pointer()).Name()
	assert.Contains(t, st, stuckFuncName)
}
