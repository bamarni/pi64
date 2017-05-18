package diskutil

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Label string

const (
	DOS Label = "o"
	GPT Label = "g"
	SUN Label = "s"
	SGI Label = "G"
)

type Disk struct {
	path       string
	size       int64
	label      Label
	partitions map[int]*Partition
}

func NewDisk(path string) (*Disk, error) {
	device, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer device.Close()

	size, err := ioctl_BLKGETSIZE64(device.Fd())
	if err != nil {
		return nil, err
	}

	// TODO: find and add existing partitions
	return &Disk{
		path: path,
		size: size,
	}, nil
}

func (d *Disk) Path() string {
	return d.path
}

func (d *Disk) Size() int64 {
	return d.size
}

// Label creates a label for the partition table.
func (d *Disk) Label(label Label) error {
	fdisk := exec.Command("fdisk", d.path)
	fdisk.Stdin = strings.NewReader(string(label) + "\nw\n")
	if err := fdisk.Run(); err != nil {
		return err
	}
	d.label = label
	return nil
}

func (d *Disk) CreatePartition(nb int, p *Partition) error {
	var commands bytes.Buffer
	var end string
	if p.end > 0 {
		end = strconv.Itoa(p.end)
	}
	// n p [i+1] [start] [end] t [i+1] [type]
	commands.WriteString(fmt.Sprintf("n\np\n%d\n%d\n%s\nt\n", nb, p.start, end))

	if len(d.partitions) > 0 {
		commands.WriteString(fmt.Sprintf("%d\n", nb))
	}
	commands.WriteString(fmt.Sprintf("%s\nw\n", p.ptype))

	fdisk := exec.Command("fdisk", d.path)
	fdisk.Stdin = &commands
	if err := fdisk.Run(); err != nil {
		return err
	}
	d.partitions[nb] = p
	return nil
}

func (d *Disk) DeletePartition(nb int) error {
	if _, ok := d.partitions[nb]; !ok {
		return fmt.Errorf("couldn't find partition %d", nb)
	}

	var commands bytes.Buffer
	commands.WriteString("d\n")
	if len(d.partitions) > 1 {
		commands.WriteString(fmt.Sprintf("%d\n", nb))
	}
	commands.WriteString("w\n")

	fdisk := exec.Command("fdisk", d.path)
	fdisk.Stdin = &commands
	if err := fdisk.Run(); err != nil {
		return err
	}
	delete(d.partitions, nb)
	return nil
}
