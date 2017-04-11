package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
)

const (
	_ int = iota
	CONFIG_TIMEZONE
	CONFIG_LOCALES
	CONFIG_WIFI
	CHANGE_HOSTNAME
	CHECK_CPU_FREQ
)

var commands = map[int]string{
	CONFIG_TIMEZONE: "Configure timezone",
	CONFIG_LOCALES:  "Configure locales",
	CONFIG_WIFI:     "Configure Wi-Fi",
	CHANGE_HOSTNAME: "Change hostname",
	CHECK_CPU_FREQ:  "Check CPU frequency",
}

func main() {
	if os.Getpid() == 1 {
		initSetup()
		return
	}

	if os.Geteuid() != 0 {
		fmt.Println("pi64-config must be run as root")
		os.Exit(1)
	}

	for {
		switch showMenu() {
		case CONFIG_TIMEZONE:
			configureTimezone()
		case CONFIG_LOCALES:
			configureLocales()
		case CONFIG_WIFI:
			configureWifi()
		case CHANGE_HOSTNAME:
			changeHostname()
		case CHECK_CPU_FREQ:
			checkCPUFreq()
		default:
			return
		}
	}
}

func attachCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func showMenu() int {
	args := []string{"0"}
	for i, len := 1, len(commands); i <= len; i++ {
		args = append(args, strconv.Itoa(i), commands[i])
	}
	res, _ := strconv.Atoi(showPrompt("menu", "pi64 config tool", args...))
	return res
}

func showMessage(msg string) {
	attachCommand("/usr/bin/dialog", "--msgbox", msg, "10", "80")
}

func showYesNo(msg string) bool {
	err := attachCommand("/usr/bin/dialog", "--yesno", msg, "10", "80")
	return err == nil
}

func showPrompt(kind, msg string, args ...string) string {
	var res bytes.Buffer
	args = append([]string{"--" + kind, msg, "0", "80"}, args...)

	cmd := exec.Command("/usr/bin/dialog", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = &res

	if err := cmd.Run(); err != nil {
		return ""
	}
	return res.String()
}

func configureTimezone() {
	if err := attachCommand("/usr/sbin/dpkg-reconfigure", "tzdata"); err != nil {
		fmt.Println("Couldn't configure timezone.")
		os.Exit(1)
	}
	showMessage("Timezone has been configured.")
}

func configureLocales() {
	showMessage("A list of generatable locales will be prompted.\n\nPlease use [SPACE] to select from that list.")
	if err := attachCommand("/usr/sbin/dpkg-reconfigure", "locales"); err != nil {
		fmt.Println("Couldn't configure locales.")
		os.Exit(1)
	}
	showMessage("Locales have been configured.")
}

func configureWifi() {
	showMessage("This functionality is not implemented yet.")
}

func changeHostname() {
	currentHostname, _ := ioutil.ReadFile("/etc/hostname")
	hostname := showPrompt("inputbox", "You can edit the current hostname below :", string(bytes.TrimRight(currentHostname, "\n")))
	if hostname == "" {
		showMessage("Hostname not provided, aborting.")
		return
	}
	if err := setHostname(hostname); err != nil {
		showMessage("Couldn't set hostname :" + err.Error())
		return
	}

	reboot := showYesNo("The hostname has been updated. However a reboot is required to properly finish this operation.\n\nDo you want to reboot now?")
	if reboot {
		attachCommand("/usr/bin/clear")
		fmt.Println("Rebooting now...")
		exec.Command("/sbin/shutdown", "-r", "now").Run()
	}
}

func checkCPUFreq() {
	showMessage("This functionality is not implemented yet.")
}
