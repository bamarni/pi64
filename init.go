package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

func initSetup() {
	time.Sleep(time.Second * 10)

	fmt.Println("Setting $PATH...")
	os.Setenv("PATH", "/bin:/sbin:/usr/sbin:/usr/bin")

	fmt.Println("Mounting filesystems...")
	fmt.Println(mountFilesystems())

	fmt.Println("Expanding root partition...")
	checkError(expandRootPartition())

	fmt.Println("Configuring packages (this may take a few minutes)...")
	checkError(configurePackages())

	fmt.Println("Adding user...")
	checkError(addUser())

	fmt.Println("Self-removing from init...")
	checkError(removeInit())

	fmt.Println("Installation succeeded! Rebooting in 5 seconds...")
	time.Sleep(time.Second * 5)

	syscall.Sync() // reboot(2) - LINUX_REBOOT_CMD_RESTART : If not preceded by a sync(2), data will be lost.
	checkError(syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART))
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		time.Sleep(time.Second * 15)
		os.Exit(1)
	}
}

func runCommand(cmd ...string) error {
	out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("couldn't run command '%s' :\n%s", cmd, out)
	}
	return err
}

func mountFilesystems() error {
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return fmt.Errorf("couldn't mount '/proc' : %s", err)
	}
	if err := syscall.Mount("sys", "/sys", "sysfs", 0, ""); err != nil {
		return fmt.Errorf("couldn't mount '/sys' : %s", err)
	}
	if err := syscall.Mount("/dev/mmcblk0p1", "/boot", "vfat", 0, ""); err != nil {
		return fmt.Errorf("couldn't mount '/boot' : %s", err)
	}
	if err := syscall.Mount("", "/", "", syscall.MS_REMOUNT, ""); err != nil {
		return fmt.Errorf("couldn't remount '/' : %s", err)
	}
	return nil
}

func expandRootPartition() error {
	rawSize, err := ioutil.ReadFile("/sys/block/mmcblk0/size")
	if err != nil {
		return err
	}
	size, err := strconv.Atoi(string(rawSize[:len(rawSize)-1]))
	if err != nil {
		return err
	}
	if err := runCommand("/sbin/parted", "/dev/mmcblk0", "u", "s", "resizepart", "2", strconv.Itoa(size-1)); err != nil {
		return err
	}
	return runCommand("/sbin/resize2fs", "/dev/mmcblk0p2")
}

func configurePackages() error {
	policyPath := "/usr/sbin/policy-rc.d"
	if err := ioutil.WriteFile(policyPath, []byte("exit 101\n"), 0755); err != nil {
		return err
	}

	os.Setenv("LC_ALL", "C")
	os.Setenv("LANGUAGE", "C")
	os.Setenv("LANG", "C")
	os.Setenv("DEBIAN_FRONTEND", "noninteractive")
	os.Setenv("DEBCONF_NONINTERACTIVE_SEEN", "true")

	if err := runCommand("/var/lib/dpkg/info/dash.preinst", "install"); err != nil {
		return err
	}
	if err := runCommand("/usr/bin/dpkg", "--configure", "-a"); err != nil {
		return err
	}
	return os.Remove(policyPath)
}

func addUser() error {
	if err := runCommand("/usr/sbin/useradd", "-s", "/bin/bash", "--create-home", "-p", "paI8KFtCOiEM6", "pi"); err != nil {
		return err
	}
	return ioutil.WriteFile("/etc/sudoers.d/010_pi-nopasswd", []byte("pi ALL=(ALL) NOPASSWD: ALL\n"), 0644)
}

func removeInit() error {
	path := "/boot/cmdline.txt"
	line, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	newLine := bytes.Replace(line, []byte("init=/usr/bin/pi64"), nil, 1)
	return ioutil.WriteFile(path, newLine, 0)
}
