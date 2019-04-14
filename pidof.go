// +build linux

package proc

import (
	"bufio"
	"bytes"
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

// PidOf lookup pid by exec name
// See https://github.com/golang/tools/tree/master/internal/fastwalk
func PidOf(name string) (pids []uint64, err error) {
	cmdName := []byte(name)
	fixedHdr := uint16(unsafe.Offsetof(syscall.Dirent{}.Name))
	fd, err := syscall.Open("/proc", 0, 0)
	if err != nil {
		return
	}
	defer syscall.Close(fd)

	buf := make([]byte, 8<<10)
	cur := 0
	read := 0

	for {
		if cur >= read {
			cur = 0
			read, err = syscall.ReadDirent(fd, buf)
			if err != nil {
				return
			}
			if read <= 0 {
				return
			}
		}
		entry := (*syscall.Dirent)(unsafe.Pointer(&buf[cur]))
		entryLen := read - cur
		if v := unsafe.Offsetof(entry.Reclen) + unsafe.Sizeof(entry.Reclen); uintptr(entryLen) < v {
			panic("header size is bigger than buffer")
		}
		if entryLen < int(entry.Reclen) {
			panic("record length is bigger than buffer")
		}
		if entry.Type == syscall.DT_DIR {
			nameBuf := (*[unsafe.Sizeof(entry.Name)]byte)(unsafe.Pointer(&entry.Name[0]))
			nameBufLen := uint16(len(nameBuf))
			limit := entry.Reclen - fixedHdr
			if limit > nameBufLen {
				limit = nameBufLen
			}
			nameLen := bytes.IndexByte(nameBuf[:limit], 0)
			if nameLen < 0 {
				panic("failed to find terminating 0 byte")
			}
			if nameBuf[0] != '.' {
				if nameBuf[0] >= 48 && nameBuf[0] <= 57 {
					pidStr := string(nameBuf[:nameLen])
					pid := lookUp(pidStr, cmdName)
					if pid > 0 {
						pids = append(pids, pid)
					}
				}
			}
		}
		cur += int(entry.Reclen)
	}
}

func lookUp(pid string, cmdName []byte) uint64 {
	fd, err := os.Open("/proc/" + pid + "/cmdline")
	if err != nil {
		return 0
	}
	defer fd.Close()
	rd := bufio.NewReader(fd)
	line, isPrefix, err := rd.ReadLine()
	for isPrefix || err != nil {
		return 0
	}
	if bytes.LastIndex(line, cmdName) != -1 {
		v, err := strconv.ParseUint(pid, 10, 64)
		if err != nil {
			return 0
		}
		return v
	}
	return 0
}
