package main

import (
	"github.com/bamarni/pi64/pkg/dialog"
	"github.com/bamarni/pi64/pkg/util"
)

func configureLocales() {
	dialog.Message("A list of generatable locales will be prompted.\n\nPlease use [SPACE] to select from that list.")
	if err := util.AttachCommand("/usr/sbin/dpkg-reconfigure", "locales"); err != nil {
		dialog.Message("Couldn't configure locales.")
		return
	}
	dialog.Message("Locales have been configured.")
}
