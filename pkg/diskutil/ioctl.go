package diskutil

// #include <linux/fs.h>
import "C"

import (
	"syscall"
	"unsafe"
)

func ioctl_BLKGETSIZE64(fd uintptr) (int64, error) {
	var size int64
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, C.BLKGETSIZE64, uintptr(unsafe.Pointer(&size))); err != 0 {
		return 0, err
	}
	return size, nil
}
