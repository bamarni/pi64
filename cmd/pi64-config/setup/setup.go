package setup

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strings"
	"syscall"
	"time"

	"github.com/bamarni/pi64/pkg/networking"
	"github.com/bamarni/pi64/pkg/util"
)

func Finish() {
	fmt.Println(util.Logo() + "\n\n")
	setEnv()

	fmt.Println("Mounting filesystems...")
	checkError(mountFilesystems())

	fmt.Println("Expanding root partition...")
	checkError(expandRootPartition())

	fmt.Println("Configuring packages (this takes a few minutes)...")
	checkError(configurePackages())

	fmt.Println("Setting hostname...")
	checkError(networking.SetHostname("raspberrypi"))

	if _, err := user.Lookup("pi"); err != nil {
		fmt.Println("Adding pi user...")
		checkError(addPiUser())
	}

	fmt.Println("Self-removing from init...")
	checkError(removeInit())

	if _, err := os.Stat("/boot/setup"); !os.IsNotExist(err) {
		fmt.Println("Running setup script (/boot/setup)...")
		if err := util.AttachCommand("/bin/bash", "/boot/setup"); err != nil {
			fmt.Println("Error : user-script /boot/setup didn't run successfully.")
		}
		os.Remove("/boot/setup")
	}

	fmt.Println("Installation succeeded! Rebooting in 5 seconds...")
	syscall.Sync() // reboot(2) - LINUX_REBOOT_CMD_RESTART : If not preceded by a sync(2), data will be lost.
	time.Sleep(time.Second * 5)
	checkError(syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART))
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		sig := make(chan os.Signal)
		signal.Notify(sig, syscall.SIGINT)
		for {
			select {
			case <-sig:
			}
		}
	}
}

func runCommand(cmd string, args ...string) error {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("couldn't run command %v :\n%s", append([]string{cmd}, args...), out)
	}
	return err
}

func setEnv() {
	os.Setenv("PATH", "/bin:/sbin:/usr/sbin:/usr/bin")
	os.Setenv("LC_ALL", "C")
	os.Setenv("LANGUAGE", "C")
	os.Setenv("LANG", "C")
	os.Setenv("DEBIAN_FRONTEND", "noninteractive")
	os.Setenv("DEBCONF_NONINTERACTIVE_SEEN", "true")
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
	parted := exec.Command("/sbin/parted", "/dev/mmcblk0")

	// parted: "-1s" specifies exactly the last sector.
	parted.Stdin = strings.NewReader("resizepart 2 -1s\n")

	if out, err := parted.CombinedOutput(); err != nil {
		return fmt.Errorf("couldn't expand root partition : %s\n\n%s", err, out)
	}
	return runCommand("/sbin/resize2fs", "/dev/mmcblk0p2")
}

func configurePackages() error {
	policyPath := "/usr/sbin/policy-rc.d"
	if err := ioutil.WriteFile(policyPath, []byte("exit 101\n"), 0755); err != nil {
		return err
	}

	cmd := exec.Command("/usr/bin/debconf-set-selections")
	cmd.Stdin = bytes.NewReader([]byte("locales locales/default_environment_locale      select  en_US.UTF-8\nlocales locales/locales_to_be_generated multiselect     en_US.UTF-8 UTF-8"))
	if err := cmd.Run(); err != nil {
		return err
	}

	if err := runCommand("/var/lib/dpkg/info/dash.preinst", "install"); err != nil {
		return err
	}
	if err := runCommand("/usr/bin/dpkg", "--configure", "-a"); err != nil {
		return err
	}
	return os.Remove(policyPath)
}

func addPiUser() error {
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
	newLine := bytes.Replace(line, []byte("init=/usr/bin/pi64-config"), nil, 1)
	return ioutil.WriteFile(path, newLine, 0)
}
