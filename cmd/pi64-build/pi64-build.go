package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/bamarni/pi64/pkg/util"
)

const (
	Lite    string = "lite"
	Desktop string = "desktop"
)

var (
	buildDir string
	rootDir  string
	bootDir  string
	version  string
	packages = []string{
		// Base packages
		"apt", "systemd", "systemd-sysv", "udev", "kmod", "locales", "sudo",

		// Networking packages
		"netbase", "net-tools", "ethtool", "iproute", "iputils-ping", "ifupdown", "dhcpcd5", "firmware-brcm80211", "wpasupplicant", "ssh", "avahi-daemon", "ntp",

		// Packages required by the pi64-config CLI tool
		"dialog", "sysbench", "wireless-tools", "parted",
	}
	desktopPackages = []string{"task-lxde-desktop"}
)

func main() {
	flag.StringVar(&buildDir, "build-dir", "", "Build directory")
	flag.StringVar(&version, "version", Lite, "pi64 version ('lite' or 'desktop')")
	flag.Parse()

	if version != Lite && version != Desktop {
		fmt.Fprintln(os.Stderr, "Unsupported version "+version)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Creating build directory...")
	if err := makeBuildDir(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Creating image...")
	checkError(createImage())

	fmt.Fprintln(os.Stderr, "Installing Debian...")
	checkError(installDebian())
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func makeBuildDir() error {
	var err error
	if buildDir, err = filepath.Abs(buildDir); err != nil {
		return err
	}
	return os.MkdirAll(buildDir, 0755)
}

func createImage() error {
	img := "pi64-" + version + ".img"

	file, err := os.OpenFile(buildDir+"/"+img, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	// allocate a big enough file depending on the version (1GB or 4GB, it will be shrinked anyway later on)
	byteSize := int64(1024 * 1024 * 1024)
	if version == Desktop {
		byteSize *= 4
	}
	if err := syscall.Fallocate(int(file.Fd()), 0, 0, byteSize); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "   Partitioning image...")
	fdisk := exec.Command("fdisk", buildDir+"/"+img)
	fdisk.Stdin = strings.NewReader("o\nn\np\n1\n8192\n137215\nt\nc\nn\np\n2\n137216\n\nw\n") // o n p 1 8192 137215 t c n p 2 137216 [Enter] w
	if err := fdisk.Run(); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "   Mapping partitions...")
	kpartx := exec.Command("kpartx", "-avs", buildDir+"/"+img)
	kpartxOut, err := kpartx.Output()
	if err != nil {
		return err
	}
	parts := bytes.Fields(kpartxOut)
	if len(parts) != 18 {
		return errors.New("couldn't parse kpartx output : " + string(kpartxOut))
	}
	bootDev, rootDev := "/dev/mapper/"+string(parts[2]), "/dev/mapper/"+string(parts[11])
	fmt.Print(rootDev)

	fmt.Fprintln(os.Stderr, "   Creating filesystems...")
	if err := exec.Command("mkdosfs", "-n", "boot", "-F", "32", bootDev).Run(); err != nil {
		return fmt.Errorf("couldn't make filesystem on %s : %s", bootDev, err)
	}
	if err := exec.Command("mkfs.ext4", "-O", "^huge_file", rootDev).Run(); err != nil {
		return fmt.Errorf("couldn't make filesystem on %s : %s", rootDev, err)
	}

	bootDir, rootDir = buildDir+"/boot-"+version, buildDir+"/root-"+version
	fmt.Fprintln(os.Stderr, "   Mounting filesystems...")
	os.Mkdir(bootDir, 0755)
	os.Mkdir(rootDir, 0755)
	if err := syscall.Mount(bootDev, bootDir, "vfat", 0, ""); err != nil {
		return err
	}
	return syscall.Mount(rootDev, rootDir, "ext4", 0, "")
}

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

	if err := exec.Command("rm", "-rf", "/var/lib/apt/lists/*", "/etc/apt/sources.list.d/*", "/usr/bin/qemu-aarch64-static").Run(); err != nil {
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
	return ioutil.WriteFile("/etc/wpa_supplicant/wpa_supplicant.conf", wpaSup, 0600)
}
