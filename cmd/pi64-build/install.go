package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/bamarni/pi64/pkg/util"
)

func installDebian() error {
	fmt.Fprintln(os.Stderr, "   Running multistrap...")
	multistrap := exec.Command("multistrap", "-a", "arm64", "-d", rootDir, "-f", "/dev/stdin")

	packages := packages
	if version == Desktop {
		packages = append(packages, desktopPackages...)
	}
	multistrap.Stdin = strings.NewReader(`
[General]
noauth=true
unpack=true
allowrecommends=true
debootstrap=Debian
aptsources=Debian

[Debian]
source=http://deb.debian.org/debian/
keyring=debian-archive-keyring
components=main non-free
suite=stretch
packages=` + strings.Join(packages, " "))

	if err := multistrap.Run(); err != nil {
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

	return os.Remove("/usr/bin/qemu-aarch64-static")

}
