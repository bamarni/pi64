package diskutil

import (
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

func (p *Partition) MkFs(fs FsType, flags ...string) error {
	flags = append(flags, p.path)

	err := exec.Command("mkfs."+string(fs), flags...).Run()
	if err == nil {
		p.fstype = fs
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

func (p *Partition) Unmount(target string, flags int) error {
	err := syscall.Unmount(p.mountpoint, flags)
	if err == nil {
		p.mountpoint = ""
	}
	return err
}
