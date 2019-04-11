package proc

import (
	"bufio"
	"bytes"
	"os"
	"strconv"
)

func readMem(pid uint64) (uint64, uint64, error) {
	var heap uint64
	fd, err := os.Open("/proc/" + strconv.FormatUint(pid, 10) + "/smaps")
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
