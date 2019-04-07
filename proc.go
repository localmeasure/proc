package proc

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
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

func Collect(pid uint, interval time.Duration) <-chan Stat {
	stats := make(chan Stat)
	ticker := time.Tick(interval)
	for range ticker {
		stats <- collect(pid, interval)
	}
	return stats
}

func collect(pid uint, interval time.Duration) Stat {
	var stat Stat
	done := make(chan struct{}, 2)
	ctx, cancel := context.WithTimeout(context.Background(), interval)
	defer cancel()
	go func() {
		var err error
		stat.Heap, stat.Stack, err = readHeapStack(pid)
		if err != nil {
			log.Println(err)
		}
		done <- struct{}{}
	}()
	go func() {
		var err error
		stat.Fd, err = readFd(pid)
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

func readHeapStack(pid uint) (uint64, uint64, error) {
	var heap uint64
	fd, err := os.Open("/proc/" + strconv.FormatUint(uint64(pid), 10) + "/smaps")
	if err != nil {
		return 0, 0, err
	}
	defer fd.Close()
	rd := bufio.NewReaderSize(fd, 1<<15)
	privDirt, size := false, false
	l := 1
	line, isPrefix, err := rd.ReadLine()
	for !isPrefix && err == nil {
		if l == 1 && line[len(line)-1] != ']' {
			privDirt = true
		}
		if privDirt && l == 10 && string(line[:14]) == "Private_Dirty:" {
			line = line[14:]
			line = bytes.TrimLeft(line, " ")
			i := bytes.IndexByte(line, ' ')
			kbs, err := strconv.ParseUint(string(line[:i]), 10, 64)
			if err != nil {
				return 0, 0, err
			}
			heap += kbs
			privDirt = false
		}
		if l == 1 && string(line[len(line)-7:]) == "[stack]" {
			size = true
		}
		if size && l == 2 && string(line[:5]) == "Size:" {
			line = line[5:]
			line = bytes.TrimLeft(line, " ")
			i := bytes.IndexByte(line, ' ')
			kbs, err := strconv.ParseUint(string(line[:i]), 10, 64)
			if err != nil {
				return 0, 0, err
			}
			return heap, kbs, nil
		}
		l++
		if l == 22 {
			l = 1
		}
		line, isPrefix, err = rd.ReadLine()
	}
	if isPrefix {
		return 0, 0, errBufferLimit
	}
	return 0, 0, err
}

type walker struct {
	counter uint64
}

func (w *walker) walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		atomic.AddUint64(&w.counter, 1)
	}
	return nil
}

func (w *walker) count() uint64 {
	return atomic.LoadUint64(&w.counter)
}

func readFd(pid uint) (uint64, error) {
	w := walker{}
	err := filepath.Walk("/proc/"+strconv.FormatUint(uint64(pid), 10)+"/fd", w.walk)
	if err != nil {
		return 0, err
	}
	// stdin, stdout, stderr
	return w.count() - 3, nil
}
