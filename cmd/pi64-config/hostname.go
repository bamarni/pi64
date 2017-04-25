package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"

	"github.com/bamarni/pi64/pkg/dialog"
	"github.com/bamarni/pi64/pkg/networking"
	"github.com/bamarni/pi64/pkg/util"
)

func changeHostname() {
	currentHostname, _ := ioutil.ReadFile("/etc/hostname")
	hostname := dialog.Prompt("inputbox", "You can edit the current hostname below :", string(bytes.TrimRight(currentHostname, "\n")))
	if hostname == "" {
		dialog.Message("Hostname not provided, aborting.")
		return
	}
	if err := networking.SetHostname(hostname); err != nil {
		dialog.Message("Couldn't set hostname :" + err.Error())
		return
	}

	reboot := dialog.YesNo("The hostname has been updated.\n\nA reboot is required to properly finish this operation, do you want to reboot now?")
	if reboot {
		util.AttachCommand("/usr/bin/clear")
		fmt.Println("Rebooting now...")
		exec.Command("/sbin/shutdown", "-r", "now").Run()
	}
}
