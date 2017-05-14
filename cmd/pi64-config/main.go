package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/bamarni/pi64/cmd/pi64-config/setup"
	"github.com/bamarni/pi64/pkg/dialog"
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
		setup.Finish()
		return
	}

	if os.Geteuid() != 0 {
		fmt.Println("pi64-config must be run as root")
		os.Exit(1)
	}

	for {
		switch ShowMenu() {
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

func ShowMenu() int {
	args := []string{"0"}
	for i, len := 1, len(commands); i <= len; i++ {
		args = append(args, strconv.Itoa(i), commands[i])
	}
	res, _ := strconv.Atoi(dialog.Prompt("menu", "pi64 config tool", args...))
	return res
}
