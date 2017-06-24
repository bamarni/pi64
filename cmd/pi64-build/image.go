package main

import (
	"fmt"
	"os"

	"github.com/bamarni/pi64/pkg/diskutil"
)

func createImage() error {
	// create a big enough image depending on the version (1GB or 4GB, it will be shrinked anyway later on)
	byteSize := int64(1024 * 1024 * 1024)
	if version == Desktop {
		byteSize *= 4
	}

	var err error
	image, err = diskutil.CreateImage(buildDir+"/pi64-"+version+".img", byteSize)
	if err != nil {
		return err
	}

	if err := image.Label(diskutil.DOS); err != nil {
		return err
	}

	bootPart = diskutil.NewPartition(diskutil.W95_FAT32_LBA, 8192, 137215)
	if err := image.CreatePartition(1, bootPart); err != nil {
		return err
	}

	rootPart = diskutil.NewPartition(diskutil.LINUX, bootPart.End()+1, 0)
	if err := image.CreatePartition(2, rootPart); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "   Mapping partitions...")
	checkError(image.MapPartitions())

	fmt.Fprintln(os.Stderr, "   Creating filesystems...")
	if err := bootPart.MkFs(diskutil.FsVFAT, "-n", "boot", "-F", "32"); err != nil {
		return err
	}
	if err := rootPart.MkFs(diskutil.FsExt4, "-b", "4096", "-O", "^huge_file"); err != nil {
		return err
	}

	rootDir = buildDir + "/root-" + version
	bootDir = rootDir + "/boot"

	fmt.Fprintln(os.Stderr, "   Mounting filesystems...")
	if err := rootPart.Mount(rootDir, 0, ""); err != nil {
		return err
	}
	return bootPart.Mount(bootDir, 0, "")
}
