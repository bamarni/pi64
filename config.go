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
	// available config commands
	_ int = iota
	CONFIG_TIMEZONE
	CONFIG_LOCALES
	CONFIG_WIFI
	CHANGE_HOSTNAME
	CHECK_CPU_FREQ

	// vcgencmd get_throttled result
	UnderVoltage        int64 = 1
	FreqCap                   = 1 << 1
	Throttling                = 1 << 2
	UnderVoltageOccured       = 1 << 16
	FreqCapOccured            = 1 << 17
	Throttled                 = 1 << 18
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

	reboot := showYesNo("The hostname has been updated.\n\nA reboot is required to properly finish this operation, do you want to reboot now?")
	if reboot {
		attachCommand("/usr/bin/clear")
		fmt.Println("Rebooting now...")
		exec.Command("/sbin/shutdown", "-r", "now").Run()
	}
}

func checkCPUFreq() {
	attachCommand("/usr/bin/dialog", "--infobox", "Running CPU frequency test... (this should take ~5 seconds)", "10", "80")
	if err := exec.Command("/usr/bin/sysbench", "--test=cpu", "--cpu-max-prime=10000", "--num-threads=4", "run").Run(); err != nil {
		showMessage("Couldn't run benchmark.")
		return
	}

	rawThrottled, err := exec.Command("/usr/sbin/vcgencmd", "get_throttled").Output()
	len := len(rawThrottled)
	if err != nil || len < 14 {
		showMessage("Couldn't run vcgencmd.")
		return
	}
	rawThrottled = rawThrottled[12 : len-1]
	throttled, err := strconv.ParseInt(string(rawThrottled), 16, 32)
	if err != nil || len < 14 {
		showMessage("Couldn't parse throttled output : " + string(rawThrottled))
		return
	}
	if throttled != 0 {
		var reason []string
		if throttled&(UnderVoltage|UnderVoltageOccured) != 0 {
			reason = append(reason, "under-voltage")
		}
		if throttled&(FreqCap|FreqCapOccured) != 0 {
			reason = append(reason, "arm frequency capped")
		}
		if throttled&(Throttling|Throttled) != 0 {
			reason = append(reason, "throttled")
		}
		showMessage(fmt.Sprintf("Throttling occured %v, your RPI doesn't perform well under load.\n\nThis usually happens because of a suboptimal power supply cable.", reason))
		return
	}

	showMessage("Congratulations! Your RPI performs well under load.")
}
