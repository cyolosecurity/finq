# finq

this is a monitoring tool for go's finalizer go routine.
it's running in the function `runfinq` under package `runtime` in go's source code.

## why to use

there is a single go routine that is responsible for executing all finalizers.
if a finalizer function is blocking, it will block the entire finalizer go routine.
this would deny cleaning up resources and memory, and it will cause a memory leak.

## how to use

```golang
go finq.Monitor(ctx, &MonitorOpts{
    StallingInterval: time.Millisecond,
    OnComplete: func(d time.Duration) {
        log.Printf("finalizer executed after %v", d)
    },
    OnStalling: func(d time.Duration) {
        log.Printf("waiting for finalizer to execute for %v", d)
    },
    ImmediatelyTriggerGC: false,
})
```

## references

https://groups.google.com/g/golang-nuts/c/uL68-fxg2K4

https://github.com/golang/go/issues/72948

https://github.com/golang/go/issues/72949

https://github.com/golang/go/issues/72950

