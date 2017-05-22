package main

import (
	"bufio"
	"os"
	"os/exec"
	"sort"
	"strconv"

	"github.com/bamarni/pi64/pkg/dialog"
	"github.com/bamarni/pi64/pkg/networking"
	"github.com/bamarni/pi64/pkg/util"
)

func configureWifi() {
	ssids, err := networking.ScanAPs("wlan0")
	if err != nil || len(ssids) == 0 {
		dialog.Message("Couldn't scan for SSIDs.")
		return
	}

	// sort access points by signal quality
	sort.Slice(ssids, func(i, j int) bool { return ssids[i].Quality > ssids[j].Quality })

	args := []string{"0"}
	for i, ssid := range ssids {
		args = append(args, strconv.Itoa(i), ssid.Name)
	}
	res, err := strconv.Atoi(dialog.Prompt("menu", "Available Wi-Fi SSIDs", args...))
	if err != nil {
		dialog.Message("SSID not provided, aborting.")
		return
	}
	ssid := ssids[res]

	passphrase := dialog.Prompt("passwordbox", "Passphrase for "+ssid.Name)
	if err != nil {
		dialog.Message("Passphrase not provided, aborting.")
		return
	}

	util.AttachCommand("/usr/bin/dialog", "--infobox", "Configuring Wi-Fi...", "10", "80")

	configFile, err := os.OpenFile("/etc/wpa_supplicant/wpa_supplicant.conf", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		dialog.Message("Couldn't open /etc/wpa_supplicant/wpa_supplicant.conf")
		return
	}
	config := bufio.NewWriter(configFile)
	config.Write([]byte("ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev\nupdate_config=1\n"))

	cmd := exec.Command("/usr/bin/wpa_passphrase", ssid.Name, passphrase)
	cmd.Stdout = config
	if err := cmd.Run(); err != nil {
		dialog.Message("Couldn't generate passphrase.")
		return
	}
	if err := config.Flush(); err != nil {
		dialog.Message("Couldn't write /etc/wpa_supplicant/wpa_supplicant.conf")
		return
	}

	if err := networking.Ifdown("wlan0"); err != nil {
		dialog.Message("Couldn't bring wlan0 interface down : " + err.Error())
		return
	}
	if err := networking.Ifup("wlan0"); err != nil {
		dialog.Message("Couldn't bring wlan0 interface up : " + err.Error())
		return
	}

	dialog.Message("Wi-Fi has been configured.")
}
