package proc

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"
)

var (
	errBufferLimit = errors.New("exceed buffer limit")
)

type Stat struct {
	Heap  uint64
	Stack uint64
	Fd    uint64
}

func Collect(pid uint64, interval time.Duration) <-chan Stat {
	stats := make(chan Stat)
	ticker := time.Tick(interval)
	for range ticker {
		stats <- collect(pid, interval)
	}
	return stats
}

func collect(pid uint64, interval time.Duration) Stat {
	var stat Stat
	done := make(chan struct{}, 2)
	ctx, cancel := context.WithTimeout(context.Background(), interval)
	defer func() {
		cancel()
		close(done)
	}()
	go func() {
		var err error
		stat.Heap, stat.Stack, err = readMem(pid)
		if err != nil {
			log.Println(err)
		}
		done <- struct{}{}
	}()
	go func() {
		var err error
		stat.Fd, err = readDir("/proc/" + strconv.FormatUint(pid, 10) + "/fd")
		if err != nil {
			log.Println(err)
		}
		done <- struct{}{}
	}()
	select {
	case <-done:
	case <-done:
	case <-ctx.Done():
		log.Println(ctx.Err())
	}
	return stat
}
