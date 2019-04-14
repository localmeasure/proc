package proc

import (
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
	go func() {
		ticker := time.Tick(interval)
		for range ticker {
			stats <- collect(pid, interval)
		}
	}()
	return stats
}

func collect(pid uint64, interval time.Duration) Stat {
	var stat Stat

	done := make(chan struct{}, 2)
	timeout := make(chan struct{})
	go func() {
		time.Sleep(interval)
		timeout <- struct{}{}
		close(timeout)
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
		// stdin, stdout, stderr
		if stat.Fd > 3 {
			stat.Fd -= 3
		}
		done <- struct{}{}
	}()

	count := 0
	for {
		select {
		case <-done:
			count++
			if count == 2 {
				close(done)
				return stat
			}
		case <-timeout:
			return stat
		}
	}
}
