package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

const (
	_ int = iota
	CONFIG_TIMEZONE
	CONFIG_LOCALES
	CONFIG_WIFI
	CHECK_CPU_FREQ
)

var commands = map[int]string{
	CONFIG_TIMEZONE: "Configure timezone",
	CONFIG_LOCALES:  "Configure locales",
	CONFIG_WIFI:     "Configure Wi-Fi",
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
	len := len(commands)
	args := []string{"--menu", "PI64 config tool", "20", "60", strconv.Itoa(len)}
	for i := 1; i <= len; i++ {
		args = append(args, strconv.Itoa(i), commands[i])
	}

	cmd := exec.Command("/usr/bin/dialog", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	var res bytes.Buffer
	cmd.Stderr = &res
	if err := cmd.Run(); err != nil {
		return 0
	}
	action, _ := strconv.Atoi(res.String())
	return action
}

func showMessage(msg string) {
	attachCommand("/usr/bin/dialog", "--msgbox", msg, "20", "60")
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

func checkCPUFreq() {
	showMessage("This functionality is not implemented yet.")
}
