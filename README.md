# finq

This is a monitoring tool for go's finalizer go routine.
The routine is spawned by the function `runfinq` under the `runtime` package in go's source code (mfinal.go).

## Why to use

The Go runtime uses a single goroutine for all finalizer execution.
A slow-running finalizer can cause a significant delay in the execution of subsequent finalizers.
Furthermore, a blocking finalizer halts the finalizer goroutine, preventing the cleanup of resources for other objects awaiting finalization, which results in a memory leak.

## Usage

```golang
go finq.Monitor(ctx, &MonitorOpts{
    StallingInterval: time.Minute,
    OnComplete: func(d time.Duration) {
        log.Printf("finalizer executed after %v", d)
    },
    OnStalling: func(d time.Duration) {
        log.Printf("waiting for finalizer to execute for %v", d)
    },
    ImmediatelyTriggerGC: false,
})
```

## References

Discussion: https://groups.google.com/g/golang-nuts/c/uL68-fxg2K4

Related Issues
* https://github.com/golang/go/issues/72948
* https://github.com/golang/go/issues/72949
* https://github.com/golang/go/issues/72950
* https://github.com/golang/go/issues/73011
