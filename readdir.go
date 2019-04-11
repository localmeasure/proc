// +build linux darwin freebsd openbsd netbsd

package proc

import (
	"syscall"
	"unsafe"
)

const mask = uint8(syscall.DT_LNK | syscall.DT_SOCK)

func readDir(dirName string) (uint64, error) {
	counter := uint64(0)
	fd, err := syscall.Open(dirName, 0, 0)
	if err != nil {
		return 0, err
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
				return 0, err
			}
			if read <= 0 {
				return counter, nil
			}
		}
		entry := (*syscall.Dirent)(unsafe.Pointer(&buf[cur]))
		if v := unsafe.Offsetof(entry.Reclen) + unsafe.Sizeof(entry.Reclen); uintptr(read) < v {
			panic("header size is bigger than buffer")
		}
		if read < int(entry.Reclen) {
			panic("record length is bigger than buffer")
		}
		if mask&entry.Type != 0 {
			counter++
		}
		cur += int(entry.Reclen)
	}
}
