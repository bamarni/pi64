package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bamarni/pi64/pkg/diskutil"
)

func finishInstall() error {
	script := exec.Command("bash", "-sex", version, rootPart.Path())
	script.Dir = buildDir

	script.Stdin = strings.NewReader(`
version=$1
root_devmap=$2

rm -rf root-$version/var/lib/apt/lists/* root-$version/etc/apt/sources.list.d/*

rsync -a linux/ root-$version/

# https://github.com/RPi-Distro/repo/issues/51
mkdir -p root-$version/lib/firmware/brcm
wget -P root-$version/lib/firmware/brcm https://github.com/RPi-Distro/firmware-nonfree/raw/master/brcm80211/brcm/brcmfmac43430-sdio.txt
`)
	if out, err := script.CombinedOutput(); err != nil {
		fmt.Fprintln(os.Stderr, string(out))
		return err
	}

	fmt.Fprintln(os.Stderr, "   Creating /boot/cmdline.txt...")
	logLevel := 3
	if debug {
		logLevel = 7
	}
	cmdLine := fmt.Sprintf("dwc_otg.lpm_enable=0 console=serial0,115200 console=tty1 root=/dev/mmcblk0p2 rootfstype=ext4 elevator=deadline fsck.repair=yes rootwait loglevel=%d net.ifnames=0 init=/usr/bin/pi64-config", logLevel)
	if err := ioutil.WriteFile(bootDir+"/cmdline.txt", []byte(cmdLine), 0644); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "   Creating /boot/config.txt...")
	if err := ioutil.WriteFile(bootDir+"/config.txt", []byte("dtparam=audio=on"), 0644); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "   Unmounting filesystems...")
	if err := bootPart.Unmount(syscall.MNT_DETACH); err != nil {
		return err
	}
	if err := rootPart.Unmount(syscall.MNT_DETACH); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "   Shrinking root filesystem...")
	if err := runCommand("e2fsck", "-fy", rootPart.Path()); err != nil {
		return err
	}

	out, err := exec.Command("resize2fs", "-P", rootPart.Path()).Output()
	if err != nil {
		return err
	}
	match := regexp.MustCompile(`Estimated minimum size of the filesystem: (\d+)`).FindStringSubmatch(string(out))
	if match == nil {
		return fmt.Errorf("couldn't parse resize2fs output : %s", out)
	}
	if err := rootPart.ResizeFs(match[1]); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "   Removing image partition maps...")
	// occasionnaly getting "device-mapper: remove ioctl on loopXp2 failed: Device or resource busy"
	// not sure why yet, sleep a bit as a workaround for now
	time.Sleep(1 * time.Second)
	if err := image.UnmapPartitions(); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "   Shrinking root partition...")
	if err := image.DeletePartition(2); err != nil {
		return err
	}

	minRootSize, _ := strconv.Atoi(match[1])
	lastSector := minRootSize*8 + rootPart.Start() - 1

	rootPart = diskutil.NewPartition(diskutil.LINUX, rootPart.Start(), lastSector)
	if err := image.CreatePartition(2, rootPart); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "   Truncating image...")
	imgSize := int64((lastSector + 1) * 512)
	if err := image.Resize(imgSize); err != nil {
		return err
	}

	syscall.Sync()

	return nil
}
