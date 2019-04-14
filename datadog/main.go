package main

import (
	"flag"
	"log"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/localmeasure/proc"
)

const ns = "goapp."

var (
	dogstatsd = flag.String("server", "127.0.0.1:8125", "dogstatsd server, example: 127.0.0.1:8125")
	name      = flag.String("name", "service-x", "service name")
	execName  = flag.String("exec", "serve", "exec filename, example: go build -o serve .")
	interval  = flag.Duration("interval", time.Second, "collect interval, example: 1s 5s 10s")
)

func main() {
	flag.Parse()
	c, err := statsd.New(*dogstatsd,
		statsd.WithNamespace(ns),
		statsd.WithTags([]string{*name}),
		statsd.Buffered(),
		statsd.WithMaxMessagesPerPayload(30),
	)
	if err != nil {
		log.Println(err)
		return
	}

	pids, err := proc.PidOf(*execName)
	if err != nil {
		log.Println(err)
		return
	}
	for _, pid := range pids {
		go func(id uint64) {
			for stat := range proc.Collect(id, *interval) {
				c.Gauge("mem.heap", float64(stat.Heap), nil, 1)
				c.Gauge("mem.stack", float64(stat.Stack), nil, 1)
				c.Gauge("fd.open", float64(stat.Fd), nil, 1)
			}
		}(pid)
	}
	select {}
}
