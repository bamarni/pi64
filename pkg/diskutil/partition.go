package diskutil

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type PartitionType string
type FsType string

const (
	// PartitionType
	W95_FAT32_LBA PartitionType = "c"
	LINUX         PartitionType = "83"

	// FsType
	FsExt4 FsType = "ext4"
	FsVFAT FsType = "vfat"
)

type Partition struct {
	ptype      PartitionType
	start      int
	end        int
	path       string
	fstype     FsType
	mountpoint string
}

func NewPartition(ptype PartitionType, start, end int) *Partition {
	return &Partition{
		ptype: ptype,
		start: start,
		end:   end,
	}
}

func (p *Partition) Path() string {
	return p.path
}

func (p *Partition) Start() int {
	return p.start
}

func (p *Partition) End() int {
	return p.end
}

func (p *Partition) MkFs(fs FsType, flags ...string) error {
	flags = append(flags, p.path)

	if out, err := exec.Command("mkfs."+string(fs), flags...).CombinedOutput(); err != nil {
		return fmt.Errorf("mkfs.%s : %s\n(%)", fs, out, err)
	}
	p.fstype = fs
	return nil
}

// ResizeFs resizes a filesystem, the size string format depends
// on the filesystem type.
func (p *Partition) ResizeFs(size string) (err error) {
	if p.mountpoint != "" {
		return errors.New("cannot resize a mounted filesystem")
	}

	switch p.fstype {
	case FsVFAT:
		if out, failed := exec.Command("fatresize", "-s", size, p.path).CombinedOutput(); failed != nil {
			err = fmt.Errorf("fatresize : %s\n(%)", out, failed)
		}
	case FsExt4:
		if out, failed := exec.Command("resize2fs", p.path, size).CombinedOutput(); failed != nil {
			err = fmt.Errorf("resize2fs : %s\n(%)", out, failed)
		}
	}
	return err
}

func (p *Partition) Mount(target string, flags uintptr, data string) error {
	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}
	err := syscall.Mount(p.path, target, string(p.fstype), flags, data)
	if err == nil {
		p.mountpoint = target
	}
	return err
}

func (p *Partition) Unmount(flags int) error {
	err := syscall.Unmount(p.mountpoint, flags)
	if err == nil {
		p.mountpoint = ""
	}
	return err
}
