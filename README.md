# proc
Zero-overhead instrumental tool collecting info on heap, stack and open sockets of a golang program.

## Usage
Build and run this along with main go program, only support **linux** for now, example ([datadog](https://github.com/localmeasure/proc/tree/master/datadog)):
```
GOOS=linux go build -o datadog .

# serve is main program name
datadog --exec serve --name service-xyz 
```

This collector is small (3MB) and won't interfere with main runtime, it sits isolatedly to collect main program stats.

vs [runtime.ReadMemStats](https://golang.org/pkg/runtime/#ReadMemStats): great detail but it **stopTheWorld** in main runtime to collect allocator stats


vs [pprof.WriteHeapProfile](https://golang.org/pkg/runtime/pprof/#WriteHeapProfile): great tool but purely for profiling
