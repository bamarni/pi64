package diskutil

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type Image struct {
	*Disk
}

// CreateImage creates an image file with pre-allocated disk space, size is in bytes.
func CreateImage(path string, size int64) (*Image, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	if err := syscall.Fallocate(int(file.Fd()), 0, 0, size); err != nil {
		return nil, err
	}

	return &Image{
		&Disk{
			path:       path,
			size:       size,
			partitions: make(map[int]*Partition),
		},
	}, file.Close()
}

func (img *Image) MapPartitions() error {
	kpartx := exec.Command("kpartx", "-avs", img.path)

	out, err := kpartx.StdoutPipe()
	if err != nil {
		return err
	}

	if err := kpartx.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(out)
	for i := 1; scanner.Scan(); i++ {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) != 9 {
			return fmt.Errorf("expected 9 fields, got %d", len(fields))
		}
		part, ok := img.Disk.partitions[i]
		if !ok {
			return fmt.Errorf("couldn't find partition %d", i)
		}
		part.path = "/dev/mapper/" + fields[2]
	}

	return kpartx.Wait()
}

func (img *Image) Shrink() {

}
