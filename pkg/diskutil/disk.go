package diskutil

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
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

func (d *Disk) Path() string {
	return d.path
}

// Format creates the partition table for the disk.
// It assumes the disk is empty.
func (d *Disk) Format(label Label, partitions []*Partition) error {
	d.label = label
	var commands bytes.Buffer

	commands.WriteString(string(d.label) + "\n")
	for i, p := range partitions {
		// n p [i+1] [start] [end] t [i+1] [type]
		var end string
		if p.end > 0 {
			end = strconv.Itoa(p.end)
		}
		command := fmt.Sprintf("n\np\n%d\n%d\n%s\nt\n", i+1, p.start, end)
		commands.WriteString(command)

		if len(d.partitions) > 0 {
			commands.WriteString(fmt.Sprintf("%d\n", i+1))
		}
		commands.WriteString(fmt.Sprintf("%s\n", p.ptype))

		d.partitions[i+1] = p
	}
	commands.WriteString("w\n")

	fdisk := exec.Command("fdisk", d.path)
	fdisk.Stdin = &commands
	return fdisk.Run()
}
