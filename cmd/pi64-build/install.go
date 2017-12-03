package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/bamarni/pi64/pkg/multistrap"
	"github.com/bamarni/pi64/pkg/pi64"
	"github.com/bamarni/pi64/pkg/util"
)

var (
	packages = []string{
		// Base packages
		"apt", "systemd", "systemd-sysv", "udev", "kmod", "locales", "sudo",

		// Networking packages
		"netbase", "net-tools", "ethtool", "iproute", "iputils-ping", "ifupdown", "dhcpcd5", "firmware-brcm80211", "wpasupplicant", "ntp",

		// Packages required by the pi64-config CLI tool
		"dialog", "sysbench", "wireless-tools", "parted",

		// Packages required by the pi64-update CLI tool
		"ca-certificates",
	}
	litePackages    = []string{"ssh", "avahi-daemon"}
	desktopPackages = []string{"task-lxde-desktop"}
	debugPackages   = []string{"device-tree-compiler", "strace", "vim", "less"}
)

func installDebian() error {
	fmt.Fprintln(os.Stderr, "   Running multistrap...")

	packages := packages
	if version == Lite {
		packages = append(packages, litePackages...)
	} else if version == Desktop {
		packages = append(packages, desktopPackages...)
	}
	if debug {
		packages = append(packages, debugPackages...)
	}
	multistrapOpts := multistrap.Options{
		Directory:  rootDir,
		Arch:       "arm64",
		Suite:      "stretch",
		Components: []string{"main", "contrib", "non-free"},
		Packages:   packages,
	}
	if err := multistrap.Run(multistrapOpts); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "   Cleaning APT...")
	if err := exec.Command("cp", "/usr/bin/qemu-aarch64-static", rootDir+"/usr/bin/qemu-aarch64-static").Run(); err != nil {
		return err
	}

	exit, err := util.Chroot(rootDir)
	if err != nil {
		return fmt.Errorf("couldn't chroot into '%s' : %s", rootDir, err)
	}
	defer exit()

	aptClean := exec.Command("apt-get", "clean")
	aptClean.Stdin = ioutil.NopCloser(bytes.NewReader(nil))
	aptClean.Stdout, aptClean.Stderr = ioutil.Discard, ioutil.Discard
	aptClean.Dir = "/"
	if err := aptClean.Run(); err != nil {
		return fmt.Errorf("couldn't run 'apt-get clean' : %s", err)
	}

	aptSources := []byte(`
deb http://deb.debian.org/debian stretch main contrib non-free
deb-src http://deb.debian.org/debian stretch main contrib non-free

deb http://deb.debian.org/debian stretch-updates main contrib non-free
deb-src http://deb.debian.org/debian stretch-updates main contrib non-free

deb http://security.debian.org/ stretch/updates main contrib non-free
deb-src http://security.debian.org/ stretch/updates main contrib non-free
`)

	if err := ioutil.WriteFile("/etc/apt/sources.list", aptSources[1:], 0644); err != nil {
		return err
	}

	// cf. https://github.com/bamarni/pi64/issues/8
	if err := exec.Command("dpkg", "--list").Run(); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "   Configuring filesystems in /etc/fstab...")
	fstab := []byte(`
proc            /proc           proc    defaults          0       0
/dev/mmcblk0p1  /boot           vfat    defaults          0       2
/dev/mmcblk0p2  /               ext4    defaults,noatime  0       1
`)
	if err := ioutil.WriteFile("/etc/fstab", fstab[1:], 0644); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "   Setting hostname...")
	if err := ioutil.WriteFile("/etc/hostname", []byte("raspberrypi"), 0644); err != nil {
		return err
	}

	// This is just in case debugging is needed, it will be overriden later on during installation
	if err := ioutil.WriteFile("/etc/resolv.conf", []byte("8.8.8.8"), 0644); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "   Configuring network interfaces...")
	netIfaces := []byte(`
allow-hotplug eth0
iface eth0 inet manual

allow-hotplug wlan0
iface wlan0 inet manual
    wpa-conf /etc/wpa_supplicant/wpa_supplicant.conf
`)
	if err := ioutil.WriteFile("/etc/network/interfaces", netIfaces[1:], 0644); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "   Configuring wpa_supplicant...")
	wpaSup := []byte("ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev\nupdate_config=1\n")
	if err := ioutil.WriteFile("/etc/wpa_supplicant/wpa_supplicant.conf", wpaSup, 0600); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "   Writing metadata...")
	metadata := pi64.Metadata{
		Version: time.Now().Format("2006-01-02"),
	}
	if err := pi64.WriteMetadata(metadata); err != nil {
		return err
	}

	return os.Remove("/usr/bin/qemu-aarch64-static")
}
