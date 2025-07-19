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
    OnStalling: func(d time.Duration) {
        log.Println("waiting for finalizer to execute for:", d)

        // you can use this functionality to retrieve the stacktrace of the blocking routine
        trace := finq.GetFinalizerStackTrace()
        if trace != "" {
            log.Println("blocking routine trace:", trace)
        }
    },
    OnComplete: func(d time.Duration) {
        log.Println("finalizer executed after:", d)
    },
    ImmediatelyTriggerGC: false,
})
```

## References
Blogpost: [Leak and Seek: A Go Runtime Mystery](https://cyolo.io/blog/leak-and-seek-a-go-runtime-mystery) ([Medium](https://medium.com/itnext/leak-and-seek-a-go-runtime-mystery-a3ac0676f0a9))

Discussion: [golang-nuts](https://groups.google.com/g/golang-nuts/c/uL68-fxg2K4)

Related Issues
* https://github.com/golang/go/issues/72948
* https://github.com/golang/go/issues/72949
* https://github.com/golang/go/issues/72950
* https://github.com/golang/go/issues/73011
